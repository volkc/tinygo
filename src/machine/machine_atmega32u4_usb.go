//go:build avr && atmega32u4
// +build avr,atmega32u4

package machine

import (
    "device/avr"
	"machine/usb"
    "runtime/interrupt"
    "unsafe"
)

// Configure the USB peripheral. The config is here for compatibility with the UART interface.
func (dev *USBDevice) Configure(config UARTConfig) {
    if dev.initcomplete {
        return
    }   

    //https://www.avrfreaks.net/forum/usb-initialization-problem
    state := interrupt.Disable()
    defer interrupt.Restore(state)

    // reset USB controller and all its registers to the default values
    avr.USBCON.ClearBits(avr.USBCON_USBE)

    // Power-On USB pads regulator
    avr.UHWCON.SetBits(avr.UHWCON_UVREGE) 

    // Configure PLL interface
    if CPUFrequency() >= 16000000 { // FIXME? does this come from UARTConfig?
        avr.PLLCSR.SetBits(avr.PLLCSR_PINDIV) // use a 16Mhz clock source
        avr.PLLFRQ.Set(avr.PLLFRQ_PLLUSB | avr.PLLFRQ_PDIV3 | avr.PLLFRQ_PDIV1) // run at 96MHz
    }

    // Enable PLL
    avr.PLLCSR.SetBits(avr.PLLCSR_PLLE)

    // Check PLL lock
    for {
        if avr.PLLCSR.HasBits(avr.PLLCSR_PLOCK) {break}
    }

    // Configure USB interface
    avr.USBCON.SetBits(avr.USBCON_USBE | avr.USBCON_OTGPADE) // turn on USB controller and power on usb pad
    avr.UDCON.ClearBits(avr.UDCON_LSM) // high speed USB 12MHz

    // Attach USB device
    avr.UDCON.ClearBits(avr.UDCON_DETACH)

    // enable interrupt for end of reset and start of frame'
    avr.UDIEN.SetBits(avr.UDIEN_EORSTE | avr.UDIEN_SOFE)

    interrupt.New(avr.IRQ_USB_GEN, handleUSBIRQ)
    interrupt.New(avr.IRQ_USB_COM, handleUSBIRQ)

    dev.initcomplete = true
}

func handleUSBSetAddress(setup usb.Setup) bool {
    // wait for transfer to complete
    timeout := 3000
    for !avr.UEINTX.HasBits(avr.UEINTX_TXINI) { 
        timeout--
        if timeout == 0 { 
            return true
        }   
    }   

    // last, set the device address to that requested by host
    avr.UDADDR.SetBits(setup.WValueL)
    avr.UDADDR.SetBits(avr.UDADDR_ADDEN)

	return true
}

func handleUSBIRQ(intr interrupt.Interrupt) {
    // reset all interrupt flags
    flags := avr.UDINT.Get()
    avr.UDINT.Set(0) // clear interrupts

    // handle end of reset interrupt
    if (flags & avr.UDINT_EORSTI) > 0 { 
        initEndpoint(0, usb.ENDPOINT_TYPE_CONTROL)

        // enable interrupts for ep0
        avr.UEIENX.SetBits(avr.UEIENX_RXSTPE)
    }

    // handle start of frame interrupt
    if (flags & avr.UDINT_SOFI) > 0 {
        // select EP0
        avr.UENUM.Set(uint8(0))

        if avr.UEBCLX.Get() > 0 { // fifocount
            avr.UEINTX.Set(0x3A) // FIFOCON=0 NAKINI=0 RWAL=1 NAKOUTI=1 RXSTPI=1 RXOUTI=0 STALLEDI=1 TXINI=0
        }

        // if you want to blink LED showing traffic, this would be the place...
    }

    // setup event received?
    if avr.UEINTX.HasBits(avr.UEINTX_RXSTPI) {
        // clear setup interrupt
        avr.UEINTX.ClearBits(avr.UEINTX_RXSTPI | avr.UEINTX_RXOUTI | avr.UEINTX_TXINI)

        b := []byte{}
        for i := 0; i < 8; i++ {
            b = append(b, avr.UEDATX.Get())
        }

        usb.NewSetup(b)
    }

    // TODO usbTxHandler
    // TODO usbRxHandler
}

// zero length packet on channel 0
func SendZlp() {
    sendUSBPacket(0, []byte{}, 0)
}

// init an endpoint // TODO
func initEndpoint(ep, config uint32) {
    if ep > 6 {
        return
    }

    // select endpoint ep
    avr.UENUM.Set(uint8(ep))

    // enable endpoint
    avr.UECONX.SetBits(avr.UECONX_EPEN) //FIXME not if config == usb.ENDPOINT_TYPE_DISABLE

    switch config {
    case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointIn:
        //TODO
        break
    case usb.ENDPOINT_TYPE_INTERRUPT | usb.EndpointOut:
        //TODO
        break
    case usb.ENDPOINT_TYPE_BULK | usb.EndpointIn:
        //TODO
        break
    case usb.ENDPOINT_TYPE_BULK | usb.EndpointOut:
        //TODO
        break
    case usb.ENDPOINT_TYPE_CONTROL:
		avr.UECFG0X.Set(0) // control endpoint
		avr.UECFG1X.SetBits(avr.UECFG1X_EPSIZE1 | avr.UECFG1X_EPSIZE0 | avr.UECFG1X_ALLOC) // 64 bytes, one bank, alloc mem
		
		avr.UEINTX.Set(0) // TODO check

		// check configuration
		if !avr.UESTA0X.HasBits(avr.UESTA0X_CFGOK) {
			//error
		}

        avr.UECFG0X.Set(0) // control endpoint
			
        break
    default:
        return
    }
}

// SendUSBInPacket sends a packet for USB (interrupt in / bulk in).
func SendUSBInPacket(ep uint32, data []byte) bool {
    sendUSBPacket(ep, data, 0)
    return true
}

// on an endpoint, send data to host
func sendUSBPacket(ep uint32, data []byte, maxsize uint16) {
    count := len(data)
    if 0 < int(maxsize) && int(maxsize) < count {
        count = int(maxsize)
    }   

    if ep == 0 { 
        copy(udd_ep_control_cache_buffer[:], data[:count])
        sendViaEPIn(ep, &udd_ep_control_cache_buffer[0], count)
    } else {
        copy(udd_ep_in_cache_buffer[ep][:], data[:count])
        sendViaEPIn(ep, &udd_ep_in_cache_buffer[ep][0], count)   
    }
}

func sendViaEPIn(ep uint32, ptr *uint8, count int) {
    // select endpoint
    avr.UENUM.Set(uint8(ep))

    eptype := avr.UECFG0X.Get() & (avr.UECFG0X_EPTYPE1 | avr.UECFG0X_EPTYPE0)

    if eptype != 0 /* CONTROL EP */ {
        if avr.UEINTX.HasBits(avr.UEINTX_RWAL) {
            return // only continue if ready to send
        }
    }

    state := interrupt.Disable()
    defer interrupt.Restore(state)

    const size = unsafe.Sizeof(uint8(0))

    p := uintptr(unsafe.Pointer(ptr))
    for i := 0; i < count; i++ {
        avr.UEDATX.Set((*(*uint8)(unsafe.Pointer(p))))

        p += size
    }

    // release TX
    avr.UEINTX.ClearBits(avr.UEINTX_FIFOCON | avr.UEINTX_NAKINI | avr.UEINTX_RXOUTI | avr.UEINTX_TXINI)
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
    var b [cdcLineInfoSize]byte

    /*
    // address
    usbEndpointDescriptors[0].DeviceDescBank[0].ADDR.Set(uint32(uintptr(unsafe.Pointer(&udd_ep_out_cache_buffer[0]))))

    // set byte count to zero
    usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.ClearBits(usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask << usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos)

    // set ready for next data
    setEPSTATUSCLR(0, sam.USB_DEVICE_ENDPOINT_EPSTATUSCLR_BK0RDY)

    // Wait until OUT transfer is ready.
    timeout := 300000
    for (getEPSTATUS(0) & sam.USB_DEVICE_ENDPOINT_EPSTATUS_BK0RDY) == 0 { 
        timeout--
        if timeout == 0 { 
            return b, ErrUSBReadTimeout
        }   
    }   

    // Wait until OUT transfer is completed.
    timeout = 300000
    for (getEPINTFLAG(0) & sam.USB_DEVICE_ENDPOINT_EPINTFLAG_TRCPT0) == 0 { 
        timeout--
        if timeout == 0 { 
            return b, ErrUSBReadTimeout
        }   
    }   

    // get data
    bytesread := uint32((usbEndpointDescriptors[0].DeviceDescBank[0].PCKSIZE.Get() >>
        usb_DEVICE_PCKSIZE_BYTE_COUNT_Pos) & usb_DEVICE_PCKSIZE_BYTE_COUNT_Mask)

    if bytesread != cdcLineInfoSize {
        return b, ErrUSBBytesRead
    }*/

    copy(b[:7], udd_ep_out_cache_buffer[0][:7])

    return b, nil
}

// EnterBootloader resets the chip into the serial bootloader.
func EnterBootloader() {
}
