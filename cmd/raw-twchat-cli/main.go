package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

var (
	insecure         = flag.Bool("insecure", false, "use a plaintext connection")
	unprintables     = flag.Bool("replace-unprintables", false, "replaces unprintable characters like SOH with \\uXXXX representation")
	unprintableRegex = regexp.MustCompile("\\pC")

	chataddr     = "irc.chat.twitch.tv"
	insecurePort = "6667"
	securePort   = "6697"
	sendChan     = make(chan string)
)

func main() {
	flag.Parse()

	pass := os.Getenv("TW_OAUTH")
	if pass == "" {
		log.Println("TW_OAUTH environment variable missing")
		os.Exit(1)
	}

	nick := os.Getenv("TW_NICK")
	if nick == "" {
		log.Println("TW_NICK environment variable missing")
		nick = "ircdefault"
	}

	var conn net.Conn
	var err error
	if insecure != nil && *insecure {
		conn, err = net.Dial("tcp", net.JoinHostPort(chataddr, insecurePort))
	} else {
		conn, err = tls.Dial("tcp", net.JoinHostPort(chataddr, securePort), &tls.Config{
			ServerName: "irc.chat.twitch.tv",
			MinVersion: tls.VersionTLS12,
		})
	}
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	log.Printf("Connected to %v\n", conn.RemoteAddr())
	conn.Write([]byte(fmt.Sprintf("PASS oauth:%s\r\n", pass)))
	log.Println("< PASS oauth:***")

	conn.Write([]byte(fmt.Sprintf("NICK %s\r\n", nick)))
	log.Printf("< NICK %s\n", nick)

	conn.Write([]byte(fmt.Sprintf("CAP REQ :twitch.tv/membership twitch.tv/commands\r\n")))
	log.Println("< CAP REQ :twitch.tv/membership twitch.tv/commands")

	var (
		done = make(chan bool, 1)
		rw   = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	)
	go readconn(rw, done)
	go sender(rw)
	go readinput(conn)
	<-done
}

func readconn(rw *bufio.ReadWriter, c chan<- bool) {
	for {
		s, err := rw.ReadString('\n')
		if err == io.EOF {
			log.Println("Remote host closed the connection.")
			break
		} else if err != nil {
			log.Println(err)
			break
		}
		if unprintables != nil && *unprintables {
			log.Printf("> %s", replaceUnprintables(s))
		} else {
			log.Printf("> %s", s)
		}
		if strings.Index(s, "PING") == 0 {
			sendChan <- "PONG"
		}
	}
	c <- true
	close(c)
	close(sendChan)
}

func sender(rw *bufio.ReadWriter) {
	for s := range sendChan {
		_, err := rw.WriteString(fmt.Sprintf("%s\r\n", s))
		if err != nil {
			log.Println(err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println(err)
		}
		if unprintables != nil && *unprintables {
			log.Printf("< %s\r\n", replaceUnprintables(s))
		} else {
			log.Printf("< %s\r\n", s)
		}
	}
}

func readinput(conn net.Conn) {
	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		sendChan <- reader.Text()
	}
	conn.Close()
}

func replaceUnprintables(s string) string {
	str := ""
	for _, c := range s {
		if unprintableRegex.MatchString(string(c)) {
			str += fmt.Sprintf("\\u%04x", c)
		} else {
			str += string(c)
		}
	}
	return strings.Replace(str, "\\u000d\\u000a", "\r\n", 1)
}
