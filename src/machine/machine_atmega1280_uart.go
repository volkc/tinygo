//go:build avr && atmega1280
// +build avr,atmega1280

package machine

import (
	"device/avr"
	"runtime/interrupt"
)

// Always use UART0 as the serial output.
var DefaultUART *UART = UART0

// UART
var (
	// UART0 is the hardware serial port on the atmega32u4.
	UART0  = &_UART0
	_UART0 = UART{
		Buffer: NewRingBuffer(),

		dataReg:    avr.UDR0,
		baudRegH:   avr.UBRR0H,
		baudRegL:   avr.UBRR0L,
		statusRegA: avr.UCSR0A,
		statusRegB: avr.UCSR0B,
		statusRegC: avr.UCSR0C,

		maskRegAFramingError:      avr.UCSR0A_FE0,
		maskRegADataOverrun:       avr.UCSR0A_DOR0,
		maskRegAParityError:       avr.UCSR0A_UPE0,
		maskRegABufferReady:       avr.UCSR0A_UDRE0,
		maskRegBEnableRX:          avr.UCSR0B_RXEN0,
		maskRegBEnableTX:          avr.UCSR0B_TXEN0,
		maskRegBEnableRXInterrupt: avr.UCSR0B_RXCIE0,
		maskRegCCharacterSize:     avr.UCSR0C_UCSZ01 | avr.UCSR0C_UCSZ00,
	}
)

func init() {
	// Register the UART interrupt.
	interrupt.New(avr.IRQ_USART0_RX, _UART0.handleInterrupt)
}
