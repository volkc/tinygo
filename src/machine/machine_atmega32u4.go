//go:build avr && atmega32u4
// +build avr,atmega32u4

package machine

import (
	"device/avr"
	"runtime/volatile"
)

const (
	portB Pin = iota * 8
	portC
	portD
	portE
	portF
)

const (
	PB0 = portB + 0
	PB1 = portB + 1
	PB2 = portB + 2
	PB3 = portB + 3
	PB4 = portB + 4
	PB5 = portB + 5
	PB6 = portB + 6
	PB7 = portB + 7
	PC6 = portC + 6
	PC7 = portC + 7
	PD0 = portD + 0
	PD1 = portD + 1
	PD2 = portD + 2
	PD3 = portD + 3
	PD4 = portD + 4
	PD5 = portD + 5
	PD6 = portD + 6
	PD7 = portD + 7
	PE2 = portE + 2
	PE6 = portE + 6
	PF0 = portF + 0
	PF1 = portF + 1
	PF4 = portF + 4
	PF5 = portF + 5
	PF6 = portF + 6
	PF7 = portF + 7
)

// getPortMask returns the PORTx register and mask for the pin.
func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	switch {
	case p >= PB0 && p <= PB7: // port B
		return avr.PORTB, 1 << uint8(p-portB)
	case p >= PC6 && p <= PC7: // port C
		return avr.PORTC, 1 << uint8(p-portC)
	case p >= PD0 && p <= PD7: // port D
		return avr.PORTD, 1 << uint8(p-portD)
	case p == PE2 || p == PE6: // port E
		return avr.PORTE, 1 << uint8(p-portE)
	default: // port F
		return avr.PORTF, 1 << uint8(p-portF)
	}
}

// SPI configuration
var SPI0 = SPI{
	spcr: avr.SPCR,
	spdr: avr.SPDR,
	spsr: avr.SPSR,
	sck:  PB1,
	sdo:  PB2,
	sdi:  PB3,
	cs:   PB0}
