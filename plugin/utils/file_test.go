package utils_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/komand/plugin-sdk-go/plugin/utils"
)

func TestFileOpenReadWrite(t *testing.T) {
	filename := "/var/cache/stuff"
	data := "sup, jerks"
	os.Remove(filename) // Cleanup the last test run incase it failed or crashed

	f, err := utils.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		t.Fatal(err)
	}
	b := []byte(data)
	f.Write(b)
	f.Close()

	f, err = utils.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		t.Fatal(err)
	}
	d, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(d) != data {
		t.Fatal(string(d))
	}
	f.Close()

	os.Remove(filename) // Cleanup

}
