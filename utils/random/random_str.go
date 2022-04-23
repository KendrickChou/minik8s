package random

import (
	"math/rand"
	"time"
)

// String : generate random string with a-zA-Z0-9
func String(n int) string {
	rand.Seed(time.Now().Unix())
	var str []byte
	i := 0
	for {
		myRand := rand.Intn(62)
		if myRand < 10 {
			str = append(str, byte(int('0')+myRand))
		} else if myRand < 36 {
			str = append(str, byte(int('a')+myRand-10))
		} else {
			str = append(str, byte(int('A')+myRand-36))
		}
		i++
		if i == n {
			return string(str)
		}
	}
}
