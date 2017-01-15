#!/bin/env python
# -*- coding:utf-8 -*-

import binascii
import hashlib
import logging
import Queue
import socket
import struct
import subprocess
import sys
from time import ctime,  time, sleep
import threading
import os


#struct ph
#{
#    uint32_t len;
#    uint32_t cmd;
#    uint32_t seq;
#};
#
#
#struct chunk_upload_head_tag
#{
#    struct ph   hdr;
#    uint8_t     upnid[NIDSIZE];
#    uint8_t     fid[FIDSIZE];
#    uint64_t    filesize;
#} chunk_upload_head;
#
#struct chunk_upload_head_resp_tag
#{
#    struct ph   hdr;
#    uint8_t     upnid[NIDSIZE];
#    uint8_t     fid[FIDSIZE];
#    uint64_t    filesize;
#    uint8_t     status;
#} chunk_upload_head_resp;
#
#struct chunk_upload_body_tag
#{
#    struct ph   hdr;
#    uint8_t     fid[FIDSIZE];
#    uint32_t    realindax;
#    uint32_t    datasize;
#    uint8_t     data[1];
#} chunk_upload_body;
#
#struct chunk_upload_body_resp_tag
#{
#    struct ph   hdr;
#    uint8_t     fid[FIDSIZE];
#    uint32_t    realindax;
#    uint32_t    datasize;
#    uint8_t     status;
#} chunk_upload_body_resp;
#------------------------------------------------------------------------------


patherror = ''
pathdesc  = ''

UPLOAD_HEADER = 'header'
UPLOAD_BODY = 'body'
#------------------------------------------------------------------------------------------------        
class work_thd (threading.Thread):
    def __init__(self, threadname, taskqueue):
        threading.Thread.__init__(self, name = threadname)
        self.terminate  = False
        self.taskqueue  = taskqueue

        self.s = socket.socket(socket.AF_UNIX,  socket.SOCK_STREAM)
        self.s.connect('/opt/cdn/cdn-socket')
        self.seq = 0


    def run(self):
        while not self.terminate:
            try:
                task = self.taskqueue.get(False)
            except:
                sleep(1)
                continue
            
            self.task_proc(task)
            

    def task_proc(self, task):
        root = task.root
        filename = task.filename
        olname, ext = os.path.splitext(filename)
        fpath = os.path.join(root, filename)
        md5 = get_file_md5(fpath)
        if olname != md5:
            print('meta md5 error %s' % (fpath))
            logger.error('%s error meta %s ' % (ctime(time()), fpath))
            newpath = '%s/%s' % (patherror, filename)
            os.rename(fpath, newpath)
            return 
    
        self.uploadfile(task)


    def uploadfile(self, task):
        CMD_CHUNK_UPLOAD_HEAD = 0x00000304
        CMD_CHUNK_UPLOAD_HEAD_RESP = 0x80000304
    
        root = task.root
        filename = task.filename
        olname, ext = os.path.splitext(filename)
        if 32 != len(olname):
            return False
    
        fpath = os.path.join(root, filename)
        filesize = os.path.getsize(fpath)
        fid = binascii.unhexlify(olname)
        upheader = struct.pack('!III6s16sQ', 42, CMD_CHUNK_UPLOAD_HEAD, self.seq,
            '000000', fid, filesize)
    
        self.senddata(upheader, fpath, UPLOAD_HEADER)
    
        self.uploadbody(fid, fpath)
    
        cr = checkupload(olname, fpath)
        if 0 == cr:
            os.unlink(fpath)
        elif 255 == cr:
            self.taskqueue.put(task)
        elif 110 == cr:
            path = '%s/%s' % (pathdesc, filename)
            os.rename(fpath, path)
    
        print('%s %s' % (cr, fpath))
    
        return cr
    

    def uploadbody(self, fid, fpath):
        BLOCK_SIZE = 1024 * 16
        CMD_CHUNK_UPLOAD_BODY = 0x00000306
        CMD_CHUNK_UPLOAD_BODY_RESP = 0x80000306
        global seq
        f = file(fpath, 'rb')
        i = 0
        while True:
            b = f.read(BLOCK_SIZE)
            if not b:
                break
            readnum = len(b)
            fmt = '!III16sII%ds' % readnum
            plen = 36 + readnum
            upbody = struct.pack(fmt, plen, CMD_CHUNK_UPLOAD_BODY, self.seq, fid, i,
                readnum, b)
    
            self.senddata(upbody, fpath, UPLOAD_BODY)
    
            i += 1
            if 0 == i % 768:
                #print 'sleep ', i
                sleep(1)
    
        f.close()


    def senddata(self, data, fpath, type):
        try:
            self.s.sendall(data)
        except socket.error, e:
            print('%s except %s' % (ctime(time()), e))
            logger.info('%s except %s %s error %s' % (ctime(time()),
                type, fpath, e))
            self.s.close()
            sleep(120)
            self.s = socket.socket(socket.AF_UNIX,  socket.SOCK_STREAM)
            self.s.connect('/opt/cdn/cdn-socket')
            self.s.sendall(data)

        self.seq += 1


    
#------------------------------------------------------------------------------------------------        
class file_info:
    def __init__(self, root, filename):
        self.root = root
        self.filename = filename
        self.stat = 0; # 0 to be upload; 1 md5 error



#------------------------------------------------------------------------------
def cur_file_dir():
    path = sys.path[0]
    if os.path.isdir(path):
        return path
    elif os.path.isfile(path):
        return os.path.dirname(path)


#------------------------------------------------------------------------------
def checkupload(olname, fpath):
    path = cur_file_dir()
    i = 0
    while i < 50: 
        output = 0
        #cmd = '%s/download %s' % (path, olname)
        cmd = '%s/download %s >> download.log ' % (path, olname)
        #logger.info('%s %s' % ( time.ctime(time.time()), cmd))
        output = subprocess.call(cmd, shell=True)
        if (0 == output):
            logger.info('%s success download %s' % (ctime(time()), olname))
            break;
        elif (100 == output):
            logger.error('%s error md5sum %s %s' % (ctime(time()), olname, fpath))
            break;
            #sys.exit(-1)
        elif (110 == output):
            logger.error('%s error describe %s' % (ctime(time()), olname))
            break;
        else:
            logger.error('%s error errno %d %s' % (ctime(time()), output, olname))
            print('%s error %d errno %d %s' % (ctime(time()), i, output, olname))
            i += 1
            sleep(13)

    return output


#------------------------------------------------------------------------------


#------------------------------------------------------------------------------
#------------------------------------------------------------------------------
#------------------------------------------------------------------------------
def get_file_md5(filename):
    if not os.path.isfile(filename):
        return
    myhash = hashlib.md5()
    f = file(filename, 'rb')
    while True:
        b = f.read(4096)
        if not b:
            break
        myhash.update(b)
    f.close()
    return myhash.hexdigest().lower()


# #------------------------------------------------------------------------------
# def processfile(root, filename):
#     olname, ext = os.path.splitext(filename)
#     fpath = os.path.join(root, filename)
#     md5 = get_file_md5(fpath)
#     if olname != md5:
#         print('meta md5 error %s' % (fpath))
#         logger.error('%s error meta %s ' % (ctime(time()), fpath))
#         return MD5_ERROR 
# 
#     uploadfile(root, filename)
# 
# 
#------------------------------------------------------------------------------



#------------------------------------------------------------------------------
def main(path, concurrency):
    queuetasks = Queue.Queue(concurrency * 2)

    thds=[]
    for i in range(concurrency):
        thds.append( work_thd('down thread %d' % i, queuetasks ))
        thds[i].setDaemon(True)
        thds[i].start()

    for root, dirs, files in os.walk(path):
        for filename in files:
            fpath = os.path.join(root, filename)
            olname, ext = os.path.splitext(filename)
            if 32 != len(olname):
                continue

            print('%s fs=%s %s' % (olname, os.path.getsize(fpath), fpath))
            fi = file_info(root, filename)
            queuetasks.put(fi)

    while not queuetasks.empty():
        sleep(1)

    for thread in thds:
        thread.terminate = True
    
    for thread in thds:
        thread.join()


#------------------------------------------------------------------------------
logger = logging.getLogger()
handler = logging.FileHandler("%s/%s.log" % (cur_file_dir(), os.path.splitext(
    os.path.basename(sys.argv[0]))[0]))
logger.addHandler(handler)
logger.setLevel(logging.NOTSET)


#------------------------------------------------------------------------------
VERSION = 12
if __name__ == '__main__':
    if len(sys.argv) < 2:
        print('%s media_path' % sys.argv[0])
        sys.exit(-1)

    path = sys.argv[1]
    pos = path.rfind('/')
    prex = path[:pos+1]
    patherror = prex + 'mediadown-error'
    pathdesc  = prex + 'mediadown-desc'
    print patherror
    print pathdesc

    logger.info("BEGIN %s" % VERSION)
    try:
        os.mkdir(patherror)
    except:
        pass
    try:
        os.mkdir(pathdesc)
    except:
        pass

    concurrency = 20
    main(sys.argv[1], concurrency)
