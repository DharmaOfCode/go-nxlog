package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/unicode"
	"github.com/dustin/go-humanize"
	"golang.org/x/sys/windows/svc/mgr"
	"flag"
)

const ConfigFilePath = "C:\\Program Files (x86)\\nxlog\\conf\\nxlog.conf"
const ConfigFileUrl = "https://www.alienvault.com/documentation/resources/downloads/nxlog.conf"
const NXlogURL = "https://nxlog.co/system/files/products/files/348/"
const NXlogDefault = "nxlog-ce-2.9.1716.msi"

type WriteCounter struct {
	Total uint64
}

type State struct {
	Verbose       bool
	Endpoint	  string
}

func ParseCmdLine() *State {
	s := State{}

	flag.BoolVar(&s.Verbose, "v", false, "Verbose output")
	flag.StringVar(&s.Endpoint, "E", "", "Endpoint IP address")

	flag.Parse()

	Banner(&s)

	return nil
}

func Process(s *State){
	Ruler(s)
	if s.Verbose{
		fmt.Println("Download of nxlog Started")
	}

	err := DownloadFile(NXlogDefault, NXlogURL + NXlogDefault, s)
	if err != nil {
		panic(err)
	}

	if s.Verbose{
		fmt.Println("Download Finished")
	}

	Ruler(s)

	err = InstallNxLog(s)
	if err != nil {
		panic(err)
	}

	err = BackupConfigFiles(s)
	if err != nil {
		panic(err)
	}

	if s.Verbose{
		fmt.Println("Download of nxlog config file started")
	}

	err = DownloadFile(ConfigFilePath, ConfigFileUrl, s)
	if err != nil {
		panic(err)
	}

	err = SetEndpoint(s)
	if err != nil {
		panic(err)
	}

	err = StartService("nxlog", s)
	if err != nil {
		log.Fatalln(err)
	}

	if s.Verbose{
		fmt.Println("Succesful configuration of NXLOG. Check your SIEM to make sure you are receiving data.")
	}

	Ruler(s)
}


func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func openFile(name string) (io.Reader, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	decoder := unicode.UTF8.NewDecoder()
	return transform.NewReader(f, unicode.BOMOverride(decoder)), nil
}

func SetEndpoint(s *State) error{
	if s.Verbose{
		fmt.Println("Setting desired endpoint")
	}

	filePath := ConfigFilePath
	file, err := openFile(filePath)
	input, err := ioutil.ReadAll(file)
	info, err := os.Stat(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	perm := info.Mode()

	newContents := strings.Replace(string(input), "define OUTPUT_DESTINATION_ADDRESS usmsensoriphere", "define OUTPUT_DESTINATION_ADDRESS " + s.Endpoint, -1)

	return ioutil.WriteFile(filePath, []byte(newContents), perm)
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func BackupConfigFiles(s *State) error{
	if s.Verbose{
		fmt.Println("Creating backup of nxlog.config")
	}
	err := CopyFile(ConfigFilePath, "C:\\Program Files (x86)\\nxlog\\conf\\nxlog.old.conf")

	return err
}

func InstallNxLog(s *State) error {
	if s.Verbose{
		fmt.Println("Installing nxlog")
	}
	_, err := exec.Command("msiexec", "/i", NXlogDefault, "/quiet").
		Output()

	return err
}

func StartService(name string, s *State) error {
	if s.Verbose{
		fmt.Println("Starting nxlog service")
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}

	defer m.Disconnect()
	service, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer service.Close()
	err = service.Start("is", "manual-started")
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

func DownloadFile(filepath string, url string, s *State) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	counter := &WriteCounter{}

	if s.Verbose{
		_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	} else {
		_, err = io.Copy(out, resp.Body)
	}
	if err != nil {
		return err
	}

	if s.Verbose{
		fmt.Print("\n")
	}

	return nil
}

func Ruler(s *State) {
	if s.Verbose{
		fmt.Println("==============================================================")
	}
}

func Banner(state *State) {
	if !state.Verbose {
		return
	}

	fmt.Println("")
	fmt.Println("go-nxlog			By Alex Useche")
	Ruler(state)
}

func main() {
	state := ParseCmdLine()
	if state != nil {
		Process(state)
	}
}
