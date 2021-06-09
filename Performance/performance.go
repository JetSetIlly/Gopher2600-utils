// This file is part of Gopher2600.
//
// Gopher2600 is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gopher2600 is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gopher2600.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/jetsetilly/gopher2600/cartridgeloader"
	"github.com/jetsetilly/gopher2600/hardware/television"
	"github.com/jetsetilly/gopher2600/hardware/television/signal"
	"github.com/jetsetilly/gopher2600/modalflag"
	"github.com/jetsetilly/gopher2600/performance"
)

const defaultInitScript = "debuggerInit"

func main() {
	runtime.GOMAXPROCS(1)

	fmt.Printf("compiled with %s\n", runtime.Version())

	md := modalflag.Modes{Output: os.Stdout}
	md.AdditionalHelp("Single command line argument to specify the ROM file")
	md.MinMax(1, 1)
	md.NewArgs(os.Args[1:])
	p, err := md.Parse()
	switch p {
	case modalflag.ParseHelp:
		return
	case modalflag.ParseTooFewArgs:
		fmt.Println("too few arguments")
		os.Exit(10)
	case modalflag.ParseTooManyArgs:
		fmt.Println("too many arguments")
		os.Exit(10)
	case modalflag.ParseError:
		fmt.Println(err)
		os.Exit(10)
	}

	romFile := md.GetArg(0)

	cartload := cartridgeloader.NewLoader(romFile, "AUTO")
	defer cartload.Close()

	tv, err := television.NewTelevision("NTSC")
	if err != nil {
		fmt.Println(err)
		os.Exit(10)
	}
	defer tv.End()

	m := &monitor{tv: tv}

	tv.SetFPSCap(false)
	tv.AddFrameTrigger(m)

	// run performance check
	err = performance.Check(os.Stdout, performance.ProfileCPU, true, tv, nil, cartload, "20s")
	if err != nil {
		fmt.Println(err)
		os.Exit(10)
	}

	fmt.Printf("avg=%f\n", m.sum/m.n)
}

type monitor struct {
	tv  *television.Television
	sum float32
	n   float32
}

func (m *monitor) NewFrame(_ television.FrameInfo) error {
	if m.tv.GetState(signal.ReqFramenum)%100 == 0 {
		fps, _ := m.tv.GetActualFPS()
		if fps < 50 {
			return nil
		}
		m.sum += fps
		m.n++
		fmt.Println(fps)
	}
	return nil
}
