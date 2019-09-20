package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	flag "github.com/spf13/pflag"
)

var (
	s, e, lines      int    = 1, 1, 72
	input_file, pipe string = "", ""
	changeLine       bool   = false
	err              error
	out              io.WriteCloser
)

func main() {
	err = getArgs()
	if err == nil {
		out = os.Stdout
		if pipe != "" {
			sendToPrinter()
		} else {
			err = readFile(input_file)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func getArgs() error {
	flag.IntVar(&s, "s", 1, "start page")
	flag.IntVar(&e, "e", 1, "end page(e>=s)")
	flag.IntVar(&lines, "1", 72, "page length")
	flag.BoolVar(&changeLine, "f", false, "change page by symbol '\\f'")
	flag.Lookup("f").NoOptDefVal = "true"
	flag.StringVar(&pipe, "d", "", "printer")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Format: ./selpg[-s n] [-e n] [-f | -l n] [-d p] [filename] [other options]")
		flag.PrintDefaults()
	}
	flag.Parse()

	arg_len := len(os.Args)
	if arg_len < 3 {
		err = errors.New("ERROR:fewer args")
		flag.Usage()
		return err
	}

	if s < 1 || e < 1 || s > 0x7fffffff || e > 0x7fffffff {
		err = errors.New("ERROR:pages should be between 1 and 2147483647")
		flag.Usage()
		return err
	}

	if s > e {
		err = errors.New("ERROR:start page should be less than end page")
		flag.Usage()
		return err
	}

	input_file = ""

	if flag.NArg() == 1 {
		_, file_err := os.Stat(flag.Args()[0])
		if file_err != nil {
			return file_err
		}
		input_file = flag.Args()[0]
	} else if flag.NArg() > 1 {
		flag.Usage()
		err = errors.New("ERROR:too many args")
	}

	if err != nil {
		return err
	}
	return nil
}

func readFile(filename string) error {
	f, err := os.Open(filename)
	if filename == "" {
		f = os.Stdin
	} else {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	defer f.Close()
	buf := bufio.NewReader(f)

	var count, page int = 0, 1
	if !changeLine {
		for page = 1; page <= e; page += 1 {
			count = 1
			for {
				if count > lines {
					break
				}
				line, err := buf.ReadString('\n')
				count += 1
				if page < s {
					if err != nil {
						return err
					} else {
						continue
					}
				}
				if err == nil {
					out.Write([]byte(line))
				} else if err == io.EOF {
					out.Write([]byte(line))
				} else {
					return err
				}
			}
		}
	} else {
		page = 0
		for page < e {
			line, err := buf.ReadString('\f')
			page += 1
			if page < s {
				if err != nil {
					return err
				} else {
					continue
				}
			}
			if err == nil {
				out.Write([]byte(line))
			} else if err == io.EOF {
				out.Write([]byte(line))
			} else {
				return err
			}
		}
	}
	return nil
}

func sendToPrinter() {
	if pipe != "" {
		cmd := exec.Command("lp", "-d", pipe)
		out, err = cmd.StdinPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		err = readFile(input_file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		out.Close()

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err = cmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		timer := time.AfterFunc(5*time.Second, func() {
			cmd.Process.Kill()
		})
		err = cmd.Wait()
		timer.Stop()
	}
}
