// Package cache abstracts away the notion of where plugins cache information into
// a set of simple function calls. Currently, the cache is a wrapper around the file
// system in /var/cache, however if you need to directly interact with the filesystem
// to more easily integrate with another library, you are free to do so. You can then
// consider this a reference for how to do so correctly.
package cache

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const cacheDir = "/var/cache/"
const lockDir = "/var/cache/lock/"
const filePerms = 0600

// InvalidCacheFileName is returned when a cache file has an invalid name
type InvalidCacheFileName string

// Error implements the error interface
func (e InvalidCacheFileName) Error() string {
	return string(e)
}

// OpenCacheFile will load the provided file from /var/cache/* and return a pointer to the
// file if found, or an error if not found / something went wrong when opening. the name
// argument should not begin with a slash, and should assume it will be appended to /var/cache
// The caller is responsible for closing the file. If they don't, there could be problems.
func OpenCacheFile(name string) (*os.File, error) {
	if err := isReservedName(name); err != nil {
		return nil, err
	}

	return openFile(cacheDir + stripLeftSlash(name))
}

// RemoveCacheFile will delete the provided file from /var/cache/* and an error if something went wrong
// the name argument should not begin with a slash, and should assume it will be appended to /var/cache
func RemoveCacheFile(name string) error {
	if err := isReservedName(name); err != nil {
		return err
	}

	return os.Remove(cacheDir + stripLeftSlash(name))
}

// CheckCacheFile checks if the file exists in the cache or not
func CheckCacheFile(name string) (bool, error) {
	return doesExist(cacheDir + stripLeftSlash(name))
}

// LockCacheFile will lock the provided file from /var/cache/* and return a boolean if the operation
// was successful or not. In the event it was not, an error may or may not be returned (always check the value first
// to know if it worked)
// the name argument should not begin with a slash, and should assume it will be appended to /var/cache
func LockCacheFile(name string) (bool, error) {
	name = lockDir + stripLeftSlash(name)
	var ok bool
	var err error
	for {
		// Spin wait until something errors, or the file becomes free
		if ok, err = doesExist(name); ok && err == nil {
			// Let's give the thread a nap while we wait, instead of pegging the CPU
			time.Sleep(1 * time.Microsecond) // TODO should this be configurable?
			continue                         // loop back to the top, try again
		}
		// If we got by the if clause because of an error, something is wrong - bail out
		if err != nil {
			return false, err
		}
		// attempt an exclusive lock - if something already grabbed the file out from under us, we simply go back to waiting
		var f *os.File
		if f, err = openExclusiveFile(name); err != nil {
			if os.IsExist(err) {
				continue // The error was that the file existed - so we just keep on a'rollin
			}
			if err != nil {
				// This was another error, some legitimate problem went wrong
				return false, err
			}
		}
		f.Close()
		break // if it ever actually gets to the end of the for loop, it means we got the exclusive lock
	}
	// If we got here, we got the lock
	return true, nil
}

// UnlockCacheFile will unlock the provided file from /var/cache/* and return a boolean if the operation
// was successful or not. In the event it was not, an error may or may not be returned (always check the value first
// to know if it worked)
// the name argument should not begin with a slash, and should assume it will be appended to /var/cache
// the timeout is used to mimic rate limiting - you can put an artificial pause on the current thread before it unlocks
// this will also keep any invocations of the process from obtaining the lock until it expires.
func UnlockCacheFile(name string, timeout *time.Duration) (bool, error) {
	// If a timeout was provided, we'll sleep for that long before unlocking the file
	// this is a very rudimentary rate-limiting mechanism
	if timeout != nil {
		time.Sleep(*timeout)
	}
	if err := os.Remove(lockDir + stripLeftSlash(name)); err != nil {
		return false, err
	}
	return true, nil
}

// We told them not to, but just incase they did, strip any leading slashes from the name arguments
func stripLeftSlash(name string) string {
	return strings.TrimLeft(name, "/")
}

func isReservedName(name string) error {
	if strings.HasSuffix(name, "/lock") {
		return InvalidCacheFileName("'lock' is a reserved name in the cache, please choose a different file name")
	}
	return nil
}

// Does not exist returns an error because there is a slight chance something goes wrong
// when checking the file system, and so the caller may need to bubble the error up to it's caller
// I'm not 100% sure how or why this might happen, but it's better to not ignore it, even at the cost
// of a slightly dumber API
func doesExist(name string) (bool, error) {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func openFile(name string) (*os.File, error) {
	return open(name, os.O_RDWR|os.O_CREATE)
}

func openExclusiveFile(name string) (*os.File, error) {
	return open(name, os.O_RDWR|os.O_CREATE|os.O_EXCL)
}

func open(name string, flags int) (*os.File, error) {
	if _, err := os.Stat(name); err != nil {
		// If it didn't exist
		var ok bool
		if ok, err = doesExist(name); !ok && err == nil {
			// Create the file and all directories leading up to it, if it didn't exist
			if err = os.MkdirAll(filepath.Dir(name), os.ModePerm); err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
	}
	// Make the file
	return os.OpenFile(name, flags, filePerms)
}
