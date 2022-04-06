package config

import (
	"errors"
	"net"
	"regexp"
	"strconv"
)

var (
	HelpRe    = regexp.MustCompile("\\?|--?he?l?p?|help")
	IPv4Re    = regexp.MustCompile("ipv4")
	IPv6Re    = regexp.MustCompile("ipv6")
	CountRe   = regexp.MustCompile("co?u?n?t?")
	SizeRe    = regexp.MustCompile("si?z?e?")
	PayloadRe = regexp.MustCompile("pa?y?l?o?a?d?")
	TTLRe     = regexp.MustCompile("ttl?")
	TimeoutRe = regexp.MustCompile("tim?e?o?u?t?")
)

type Config struct {
	IPVersion   int
	DestIP      string
	Count       int
	Size        int
	Payload     []byte
	TTL         int
	TimeoutSecs int
	HelpMode    bool
}

func Default() *Config {
	return &Config{
		IPVersion:   4,
		DestIP:      "",
		Count:       5,
		Size:        100,
		Payload:     []byte{0xAB, 0xCD},
		TTL:         64,
		TimeoutSecs: 2,
		HelpMode:    false,
	}
}

func (c *Config) ParseArgs(args []string) error {
	pos := 0

	for pos < len(args) {
		cur := args[pos]

		if HelpRe.MatchString(cur) {

			// ... help ...

			c.HelpMode = true

		} else if CountRe.MatchString(cur) {

			// ... count <num> ...

			countStr, err := GetNext(args, pos)
			if err != nil {
				return err
			}
			c.Count, err = strconv.Atoi(countStr)
			if err != nil {
				return err
			}

			pos += 1

		} else if SizeRe.MatchString(cur) {

			// ... size <num> ...

			sizeStr, err := GetNext(args, pos)
			if err != nil {
				return err
			}
			c.Size, err = strconv.Atoi(sizeStr)
			if err != nil {
				return err
			}

			pos += 1

		} else if IPv4Re.MatchString(cur) {

			// ... ipv4 ...

			c.IPVersion = 4

		} else if IPv6Re.MatchString(cur) {

			// ... ipv6 ...

			c.IPVersion = 6

		} else if PayloadRe.MatchString(cur) {

			// ... payload <string> ...

			payload, err := GetNext(args, pos)
			if err != nil {
				return err
			}
			c.Payload = []byte(payload)

			pos += 1

		} else if TTLRe.MatchString(cur) {

			// ... ttl <num> ...

			ttl, err := GetNext(args, pos)
			if err != nil {
				return err
			}
			c.TTL, err = strconv.Atoi(ttl)
			if err != nil {
				return err
			}

			pos += 1

		} else if TimeoutRe.MatchString(cur) {

			// ... timeout <num> ...

			timeout, err := GetNext(args, pos)
			if err != nil {
				return err
			}
			c.TimeoutSecs, err = strconv.Atoi(timeout)
			if err != nil {
				return err
			}

			pos += 1
		}

		pos += 1
	}

	// assume the last item is the ip
	c.DestIP = args[len(args)-1]

	return nil
}

func (c *Config) ValidIP() bool {
	ip := net.ParseIP(c.DestIP)
	if ip == nil {
		return false
	}

	if c.IPVersion == 4 && ip.To4() == nil {
		return false
	}

	return true
}

func GetNext(arr []string, index int) (string, error) {
	if index+1 > len(arr) {
		return "", errors.New("not enough arguments")
	}

	return arr[index+1], nil
}
