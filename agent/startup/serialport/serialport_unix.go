// Copyright 2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the
// License is located at
//
// http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing
// permissions and limitations under the License.
//
// +build linux

// Package serialport implements serial port capabilities
package serialport

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/aws/amazon-ssm-agent/agent/log"
)

const (
	comport = "/dev/ttyS0"
)

type SerialPort struct {
	log        log.T
	fileHandle *os.File
}

// NewSerialPort creates a serial port object with predefined parameters.
func NewSerialPort(log log.T) (sp *SerialPort) {
	return &SerialPort{
		log:        log,
		fileHandle: nil,
	}
}

// OpenPort opens the serial port which MUST be done before WritePort is called.
func (sp *SerialPort) OpenPort() (err error) {
	fileHandle, err := os.OpenFile(comport, syscall.O_RDWR, 0)

	if err != nil {
		sp.log.Errorf("Error occurred while opening serial port: %v", err.Error())
		return err
	}

	baudRate := uint32(syscall.B115200)
	state := syscall.Termios{
		Cflag:  syscall.CS8,
		Oflag:  0,
		Ospeed: baudRate,
	}

	fd := fileHandle.Fd()
	if _, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&state)),
		0,
		0,
		0,
	); err != 0 {
		sp.log.Errorf("Error occurred while configuring serial port: %v", err.Error())
		return err
	}

	sp.fileHandle = fileHandle

	return nil
}

// ClosePort closes the serial port, which MUST be done at the end.
func (sp *SerialPort) ClosePort() {
	if sp.fileHandle == nil {
		sp.log.Error("Error occurred while closing serial port: Port must be opened")
	}
	sp.fileHandle.Close()
	return
}

// WritePort writes messages to serial port, which is then picked up by ICD in EC2 droplet
// and sent to system log in console.
func (sp *SerialPort) WritePort(message string) {
	sp.log.Infof("Write to serial port: %v", message)
	formattedMessage := fmt.Sprintf("%v: %v\n", time.Now().UTC().Format("2006/01/02 15:04:05Z"), message)
	if _, err := sp.fileHandle.WriteString(formattedMessage); err != nil {
		sp.log.Errorf("Error occurred while writing to serial port: %v", err.Error())
	}

	return
}
