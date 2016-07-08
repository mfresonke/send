package main

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	runNGROK(7070)
	time.Sleep(15 * time.Second)
}

func runNGROK(port int) (tunnelURL string, err error) {
	cmd := exec.Command("ngrok", "http", strconv.Itoa(port))
	output, err := cmd.StdoutPipe()
	cmd.Start()
	check(err)
	go func() {
		err := cmd.Wait()
		check(err)
	}()
	go func(output io.ReadCloser) {
		time.Sleep(10 * time.Second)
		readme := make([]byte, 10)
		num, err := output.Read(readme)
		check(err)
		fmt.Println("read ", num, " bytes")
		fmt.Println(string(readme))
	}(output)

	// make a request to the ngrok api to check the URL
	// resp, err := http.Get("http://127.0.0.1:4040/api/tunnels")
	// check(err)
	// jsonDec := json.NewDecoder(resp.Body)
	// check(err)
	// var res struct {
	// 	Tunnels []struct {
	// 		URL string `json:"public_url"`
	// 	} `json:"tunnels"`
	// }
	// jsonDec.Decode(&res)
	// if len(res.Tunnels) != 1 {
	// 	ErrPrintln("Error, tunnel not detected in ngrok.")
	// 	os.Exit(1)
	// }
	return "", nil
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
