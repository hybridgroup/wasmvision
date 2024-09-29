package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/wasmvision/wasmvision/capture"
	"github.com/wasmvision/wasmvision/engine"
	"github.com/wasmvision/wasmvision/runtime"
)

func run(cCtx *cli.Context) error {
	processors := cCtx.StringSlice("processor")
	if len(processors) == 0 {
		fmt.Println("No wasm processors specified")
		os.Exit(1)
	}

	device := cCtx.String("device")
	mjpeg := cCtx.Bool("mjpeg")
	mjpegPort := cCtx.String("mjpegport")
	clear := cCtx.Bool("clear")
	modelsDir := cCtx.String("models-dir")
	if modelsDir == "" {
		modelsDir = DefaultModelPath()
	}

	ctx := context.Background()

	// load wasm runtime
	r := runtime.New(ctx, runtime.InterpreterConfig{ModelsDir: modelsDir})
	defer r.Close(ctx)

	for _, p := range processors {
		module, err := os.ReadFile(p)
		if err != nil {
			log.Panicf("failed to read wasm processor module: %v\n", err)
		}

		fmt.Printf("Loading wasmCV guest module %s...\n", p)
		if err := r.RegisterGuestModule(ctx, module); err != nil {
			log.Panicf("failed to load wasm processor module: %v\n", err)
		}
	}

	// Open the webcam.
	webcam := capture.NewWebcam(device)
	defer webcam.Close()
	if err := webcam.Open(); err != nil {
		log.Panicf("Error opening video capture device: %v\n", device)
	}

	var mjpegstream engine.MJPEGStream
	if mjpeg {
		mjpegstream = engine.NewMJPEGStream(mjpegPort)

		go mjpegstream.Start()
	}

	fmt.Printf("Start reading device: %v\n", device)
	i := 0

	for {
		frame, err := webcam.Read()
		if err != nil {
			fmt.Printf("frame error %v\n", device)
			frame.Close()
			continue
		}

		if frame.Empty() {
			frame.Close()
			continue
		}

		r.FrameCache.Set(frame)

		if clear {
			fmt.Print("\033[2J\033[3J\033[H")
		}

		i++
		fmt.Printf("Read frame %d\n", i+1)

		frame = r.Process(ctx, frame)

		if mjpeg {
			mjpegstream.Publish(frame)
		}

		// cleanup frame
		frame.Close()
		r.FrameCache.Delete(frame.ID)
	}
}

func DefaultModelPath() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	return filepath.Join(dirname, "models")
}
