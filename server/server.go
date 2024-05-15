package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/olahol/melody"
)

type Message struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type Output struct {
	Result string `json:"result"`
}

func runCommand(name string, command []string) string {
	c := exec.Command(name, command...)
	f, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)

	io.Copy(buf, f)

	output := buf.String()

	return output
}

func main() {
	m := melody.New()

	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		err := m.HandleRequest(w, r)
		if err != nil {
			fmt.Println(err)
		}
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		message := new(Message)
		err := json.Unmarshal(msg, message)
		if err != nil {
			fmt.Println("Error unmarshalling message:", err)
			return
		}

		fmt.Println("Received command:", message.Command)
		splitInput := strings.Fields(message.Command)

		output := runCommand(splitInput[0], splitInput[1:])

		jsonOp := Output{
			Result: output,
		}

		jsonResponse, err := json.Marshal(jsonOp)

		if err != nil {
			panic(err)
		}
		m.Broadcast(jsonResponse)

	})

	fmt.Println("Server started on :3000")
	http.ListenAndServe(":3000", nil)
}
