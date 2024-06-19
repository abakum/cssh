package main

/*
Как собрать:
`git clone https://github.com/abakum/cssh`
`cd cssh`
`go install`

Для чего:
Чтоб запускать как socks5 прокси
Например прокси от https://www.vpnjantit.com/free-ssh
создай в `~/.ssh/config` алиас `Host cssh`
```
Host cssh
 User foo-vpnjantit.com
 HostName bar.vpnjantit.com
 SessionType none
 DynamicForward 127.0.0.1:1080
 PubkeyAuthentication no
 UserKnownHostsFile ~/.ssh/bar
 LogLevel debug1
```
запусти `cssh -password=123`
видишь `encPassword ....`
допиши `encPassword ....` после `Host cssh`
в самый верх `~/.ssh/config` пиши `IgnoreUnknown *` чтоб ssh не ругался
И что этот `~/.ssh/config` если попадёт к другому раскодирует `encPassword ....` в `123`? - Да раскодирует!

Поэтому переименуй `cssh` например в `secret` (никому не говори) и переименуй алиас в `Host secret`
запусти `secret -password=123`
замени `encPassword ....` после  `Host secret`
потом просто запускай `secret`
*/

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	rdebug "runtime/debug"
	"strings"
	"time"

	"github.com/abakum/menu"
	"github.com/abakum/winssh"
	"github.com/trzsz/go-arg"

	. "github.com/abakum/cssh/tssh"
	version "github.com/abakum/version/lib"
	"github.com/xlab/closer"
	"golang.org/x/crypto/ssh"
)

type Parser struct {
	*arg.Parser
}

func (p *Parser) WriteHelp(w io.Writer) {
	var b bytes.Buffer
	p.Parser.WriteHelp(&b)
	s := strings.Replace(b.String(), "  -v, --version          show program's version number and exit\n", "", 1)
	fmt.Fprint(w, s)

}

func NewParser(config arg.Config, dests ...interface{}) (*Parser, error) {
	p, err := arg.NewParser(config, dests...)
	return &Parser{p}, err
}

const (
	TOW = time.Second * 5 //watch TO
)

var (
	_    = version.Ver
	args SshArgs
	Std  = menu.Std
	repo = base() // Имя репозитория `cssh`
	imag string   // Имя исполняемого файла `cssh` оно же имя алиаса. Можно изменить чтоб не указывать имя алиаса.
)

//go:generate go run github.com/abakum/version

//go:embed VERSION
var Ver string

func main() {
	SetColor()

	// tssh
	DebugPrefix = l.Prefix()
	DebugF = func(format string) string {
		return fmt.Sprintf("%s%s %s\r\n", l.Prefix(), src(9), format)
	}
	WarningF = func(format string) string {
		return fmt.Sprintf("%s%s %s\r\n", le.Prefix(), src(9), format)
	}

	exe, err := os.Executable()
	Fatal(err)
	imag = strings.Split(filepath.Base(exe), ".")[0]

	defer closer.Close()
	closer.Bind(cleanup)

	// Like `parser := arg.MustParse(&args)` but override built in option `-v, --version` of package `arg`
	parser, err := NewParser(arg.Config{}, &args)
	Fatal(err)

	a2s := make([]string, 0) // without built in option
	deb := false
	for _, arg := range os.Args[1:] {
		switch strings.ToLower(arg) {
		case "-help", "--help":
			parser.WriteHelp(Std)
			return
		case "-h":
			parser.WriteUsage(Std)
			return
		case "-version", "--version":
			Println(args.Version())
			return
		case "-v":
			deb = true
		default:
			a2s = append(a2s, arg)
		}
	}

	if err := parser.Parse(a2s); err != nil {
		parser.WriteUsage(Std)
		Fatal(err)
	}

	if args.Ver {
		Println(args.Version())
		return
	}
	args.Debug = args.Debug || deb

	SecretEncodeKey = append([]byte(repo+imag), make([]byte, 16)...)[:16]

	if args.Destination == "" {
		args.Destination = imag
	}
	code := TsshMain(&args)
	if args.Background {
		Println("cssh started in background with code:", code)
		closer.Hold()
	} else {
		Println("cssh exit with code:", code)
	}
}

func cleanup() {
	winssh.KidsDone(os.Getpid())
	Println("cleanup done")
	if runtime.GOOS == "windows" {
		menu.PressAnyKey("Press any key - Нажмите любую клавишу", TOW)
	}
}

func FingerprintSHA256(pubKey ssh.PublicKey) string {
	return pubKey.Type() + " " + ssh.FingerprintSHA256(pubKey)
}

func GoVer() (s string) {
	info, ok := rdebug.ReadBuildInfo()
	s = "go"
	if ok {
		s = info.GoVersion
	}
	return
}

func base() string {
	info, ok := rdebug.ReadBuildInfo()
	if ok {
		return path.Base(info.Path) //info.Main.Path
	}
	exe, err := os.Executable()
	if err == nil {
		return strings.Split(filepath.Base(exe), ".")[0]
	}
	dir, err := os.Getwd()
	if err == nil {
		return filepath.Base(dir)
	}
	return "main"
}
