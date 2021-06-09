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
//
// *** NOTE: all historical versions of this file, as found in any
// git repository, are also covered by the licence, even when this
// notice is not present ***

//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/jetsetilly/gopher2600/cartridgeloader"
	"github.com/jetsetilly/gopher2600/emulation"
	"github.com/jetsetilly/gopher2600/hardware"
	"github.com/jetsetilly/gopher2600/hardware/riot/ports"
	"github.com/jetsetilly/gopher2600/hardware/riot/ports/plugging"
)

func main() {
	worker := js.Global().Get("self")
	scr, err := NewCanvas(worker)
	if err != nil {
		scr.worker.Call("log", err.Error())
		return
	}

	// create new vcs
	vcs, err := hardware.NewVCS(scr.tv)
	if err != nil {
		scr.worker.Call("log", err.Error())
		return
	}

	// load cartridge
	cartload := cartridgeloader.Loader{
		Filename: "http://localhost:2600/example.bin",
	}

	err = vcs.AttachCartridge(cartload)
	if err != nil {
		scr.worker.Call("log", err.Error())
		return
	}

	// add message handler - implements controllers
	messageHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var err error

		data := args[0].Get("data")
		switch data.Get("cmd").String() {
		case "keyDown":
			key := data.Get("key").Int()
			switch key {
			case 37: // left
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Left, ports.DataStickTrue)
			case 39: // right
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Right, ports.DataStickTrue)
			case 38: // up
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Up, ports.DataStickTrue)
			case 40: // down
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Down, ports.DataStickTrue)
			case 32: // space
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Fire, true)
			}
		case "keyUp":
			key := data.Get("key").Int()
			switch key {
			case 37: // left
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Left, ports.DataStickFalse)
			case 39: // right
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Right, ports.DataStickFalse)
			case 38: // up
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Up, ports.DataStickFalse)
			case 40: // down
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Down, ports.DataStickFalse)
			case 32: // space
				err = vcs.RIOT.Ports.HandleEvent(plugging.LeftPlayer, ports.Fire, false)
			}
		default:
			js.Global().Get("self").Call("log", args[0].String())
		}

		if err != nil {
			panic(err)
		}

		return nil
	})
	defer func() {
		worker.Call("removeEventListener", "message", messageHandler, false)
		messageHandler.Release()
	}()
	worker.Call("addEventListener", "message", messageHandler, false)

	// run emulation
	vcs.Run(func() (emulation.State, error) {
		return emulation.Running, nil
	})
}
