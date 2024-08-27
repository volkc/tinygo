//go:build arduino_micro
// +build arduino_micro

package machine

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 16000000
}

// Digital pins.
const (
	D0  = PD2
	D1  = PD3
	D2  = PD1
	D3  = PD0
	D4  = PD4
	D5  = PC6
	D6  = PD7
	D7  = PE6
	D8  = PB4
	D9  = PB5
	D10 = PB6
	D11 = PB7
	D12 = PD6
	D13 = PC7
	D14 = PB3
	D15 = PB1
)

// LED on the Arduino
const LED Pin = D13

// ADC on the Arduino
const (
	ADC0 Pin = PF0
	ADC1 Pin = PF1
	ADC4 Pin = PF4
	ADC5 Pin = PF5
	ADC6 Pin = PF6
	ADC7 Pin = PF7
)

// UART pins
const (
	UART_TX_PIN Pin = PD3
	UART_RX_PIN Pin = PD2
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Arduino Micro"
	usb_STRING_MANUFACTURER = "Arduino LLC"
)

var (
	usb_VID uint16 = 0x2341
	usb_PID uint16 = 0x8037
)

func init() {
    initUART()
}
