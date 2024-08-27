//go:build avr && atmega32u4
// +build avr,atmega32u4

package machine

import (
	"device/avr"
	"runtime/interrupt"
)

// Always use UART0 as the serial output.
var DefaultUART *UART = UART1

// UART
var (
	// UART1 is the hardware serial port on the atmega32u4.
	UART1  = &_UART1
	_UART1 = UART{
		Buffer: NewRingBuffer(),

		dataReg:    avr.UDR1,
		baudRegH:   avr.UBRR1H,
		baudRegL:   avr.UBRR1L,
		statusRegA: avr.UCSR1A,
		statusRegB: avr.UCSR1B,
		statusRegC: avr.UCSR1C,

		maskRegAFramingError:      avr.UCSR1A_FE1,
		maskRegADataOverrun:       avr.UCSR1A_DOR1,
		maskRegAParityError:       avr.UCSR1A_UPE1,
		maskRegABufferReady:       avr.UCSR1A_UDRE1,
		maskRegBEnableRX:          avr.UCSR1B_RXEN1,
		maskRegBEnableTX:          avr.UCSR1B_TXEN1,
		maskRegBEnableRXInterrupt: avr.UCSR1B_RXCIE1,
		maskRegCCharacterSize:     avr.UCSR1C_UCSZ11 | avr.UCSR1C_UCSZ10,
	}
)

func initUART() {
	// Register the UART interrupt.
	interrupt.New(avr.IRQ_USART1_RX, _UART1.handleInterrupt)
}
