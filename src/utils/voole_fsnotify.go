// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package utils

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
)

const (
	VError = iota
	VCreate
	VWrite
	VRemove
	VRename
	VChmod
)

type FsnotifyCallback func(int, string) int

type FsnotifyMgmt struct {
	done_flag chan bool
	watcher   *fsnotify.Watcher
	cb        FsnotifyCallback
	dir       string
}

func VooleNewWatcher(dir string, cb FsnotifyCallback) (*FsnotifyMgmt, error) {

	var err error
	var mgmt FsnotifyMgmt

	mgmt.dir = dir
	mgmt.cb = cb
	mgmt.done_flag = make(chan bool)

	mgmt.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		goto exit
	}
	go event_handle(&mgmt)
	err = mgmt.watcher.Add(mgmt.dir)
	if err != nil {
		fmt.Println("watcher dir", mgmt.dir, "failed...")
		goto exit
	}
	return &mgmt, err
exit:
	close(mgmt.done_flag)
	return nil, err
}

func VooleCloseWatcher(mgmt *FsnotifyMgmt) {
	if mgmt == nil {
		return
	}
	close(mgmt.done_flag)
	if mgmt.watcher != nil {
		mgmt.watcher.Close()
	}
}

func event_handle(mgmt *FsnotifyMgmt) {
	if mgmt == nil {
		return
	}
	for {
		select {
		case _, ok := <-mgmt.done_flag:
			if ok == false {
				log.Println("done...")
				return
			}
		case event := <-mgmt.watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				mgmt.cb(VWrite, event.Name)
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				mgmt.cb(VCreate, event.Name)
			}
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				mgmt.cb(VRename, event.Name)
			}
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				mgmt.cb(VChmod, event.Name)
			}
			if event.Op&fsnotify.Remove == fsnotify.Chmod {
				mgmt.cb(VRemove, event.Name)
			}
		case <-mgmt.watcher.Errors:
			mgmt.cb(VError, "")
		}
	}
}
