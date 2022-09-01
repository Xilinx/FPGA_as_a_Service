import socket
import sys
import os
import subprocess
from threading import Thread

SERVER_IP = '10.96.59.3'
SERVER_PORT = 8010


def client_send():
    bind_ip = SERVER_IP
    bind_port = SERVER_PORT
    client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    client.connect((bind_ip, bind_port))
    client.send("fpga_sever")
    response = client.recv(4096)
    client.close()
    print "Senddd request to server..."	
    print response
    print "--END--"

def usage():
    print("Usage:")
    print("\t%s" %  sys.argv[0])

def main():
    if (len(sys.argv) != 1):
        usage()
        exit(0)

    client_send()
main()

