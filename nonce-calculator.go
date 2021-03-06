/*
   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
// CONTRIBUTORS AND COPYRIGHT HOLDERS (c) 2013:
// Dag Robøle (BM-2DAS9BAs92wLKajVy9DS1LFcDiey5dxp5c)

package main

import (
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	//"time"
)

func varint(integer uint64) []byte {

	buf := make([]byte, 16)

	if integer < 253 {

		buf[0] = byte(integer)
		return buf[:1]

	} else if integer >= 253 && integer < 65536 {

		buf[0] = 253
		binary.BigEndian.PutUint16(buf[1:], uint16(integer))
		return buf[:3]

	} else if integer >= 65536 && integer < 4294967296 {

		buf[0] = 254
		binary.BigEndian.PutUint32(buf[1:], uint32(integer))
		return buf[:5]

	} else {

		buf[0] = 255
		binary.BigEndian.PutUint64(buf[1:], uint64(integer))
		return buf[:9]
	}
}

func scan(offset_start, offset_end, target uint64, payload_hash []byte, out chan<- uint64, done chan<- bool, shutdown *bool) {

	var nonce, trials uint64 = offset_start, 18446744073709551615
	h1, h2 := sha512.New(), sha512.New()

	for trials > target {

		nonce++
		if *shutdown || nonce > offset_end {
			done <- true
			return
		}
		b := varint(nonce)
		b = append(b, payload_hash...)
		h1.Write(b)
		h2.Write(h1.Sum(nil))
		trials = binary.BigEndian.Uint64(h2.Sum(nil)[:8])
		h1.Reset()
		h2.Reset()
	}
	out <- nonce
	done <- true
}

func main() {

	payload, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println("Failed to read from stdin")
		os.Exit(1)
	}

	ncpu := runtime.NumCPU() - 1
	if ncpu < 1 {
		ncpu = 1
	}
	runtime.GOMAXPROCS(ncpu)

	sha := sha512.New()
	sha.Write(payload)
	payload_hash := sha.Sum(nil)
	var target uint64 = 18446744073709551615 / uint64((len(payload)+14000+8)*320)

	var nprocs int = 1000
	var i, slice uint64 = 0, 18446744073709551615 / uint64(nprocs)

	recv := make(chan uint64, nprocs)
	done := make(chan bool, nprocs)
	shutdown := false

	//t0 := time.Now()
	for ; i < uint64(nprocs); i++ {
		go scan(i*slice, i*slice+slice, target, payload_hash, recv, done, &shutdown)
	}
	nonce := <-recv
	//t1 := time.Now()
	//fmt.Printf("Payload size %d bytes, Nonce %d, Time %v\n", len(payload), nonce, t1.Sub(t0))
	fmt.Printf("%d", nonce)

	shutdown = true
	for i = 0; i < uint64(nprocs); i++ {
		<-done
	}
}
