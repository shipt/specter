#!/usr/bin/env python3
from ipaddress import IPv4Network, IPv4Address
from time import sleep
from random import randint
from random import getrandbits
import random
import sys

if sys.version_info[0] != 3:
    print("This script requires Python version 3")
    sys.exit(1)


# --User Input--
# input is used to read text (strings) from the user
# since Python 3.x doesn't evaluate and convert the data type, you have to explicitly convert
try:
    rps = int(input('Simulated requests per second: '))
except:
    print("The RPS input is required and should be a number")
    exit(1)
sleepTime = 1 / rps
try:
    errors = bool(input('Do you want to simulate errors? (True or False): '))
except:
    print("The Errors input required and should either True or False")
if errors == True:
    try:
        fourErrs = int(input('% of simulated 400s: '))
        if 0 <= fourErrs <= 100:
            pass
        else:
            exit(1)
    except:
        print("% of simulated 400s should be between 0 and 100")
    try:
        fiveErrs = int(input('% of simulated 500s: '))
        if 0 <= fiveErrs <= 100:
            pass
        else:
            exit(1)
    except:
        print("% of simulated 500s should be between 0 and 100")
    if fourErrs + fiveErrs > 100:
        print("total % of simulated errors is above 100. Please keep the total % of errors at 100 or lower.")
        exit(1)


# --Log message generation--
host = 'spooky.specter.local '
# remoteAddr='174.207.18.130'
remoteUser = '- '
timeLocal = '[14/Jan/2019:19:38:45 +0000] '
request = '"POST / HTTP/1.1" '
# status = '200 '
bodyBytesSent = '2105 '
httpReferer = '"http://localhost:8080/" '
httpUserAgent = '"Mozilla/5.0 (iPhone; CPU iPhone OS 12_1_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/16C104/ndbflbhxuptclhugenrdeeiocpjxvttm" '
httpXForwardedFor = '"-" '
requestID = '"foo" '
requestTime = '"0.307" '
upstreamResponseTime = '"0.296" '
upstreamConnectTime = '"0.001"'
subnets = ["24.0.0.0/8", "32.0.0.0/8"]


# --Writing the log file--
while True:
    # --IP Generation--
    subnet = IPv4Network(random.choice(subnets))

    # subnet.max_prefixlen contains 32 for IPv4 subnets and 128 for IPv6 subnets
    # subnet.prefixlen is 24 in this case, so we'll generate only 8 random bits
    bits = getrandbits(subnet.max_prefixlen - subnet.prefixlen)

    # here, we combine the subnet and the random bits
    # to get an IP address from the previously specified subnet
    addr = IPv4Address(subnet.network_address + bits)
    addr_str = str(addr)
    # have to add a space at the end for log formatting
    remoteAddr = addr_str + " "

    # --Logfile Generation--
    twohundreds = 100 - (fourErrs + fiveErrs)
    statusList = [200] * twohundreds + [404] * fourErrs + [503] * fiveErrs
    status = random.choice(statusList)
    log = host + remoteAddr + "- " + remoteUser + timeLocal + request + str(status) + " " + bodyBytesSent + httpReferer + \
        httpUserAgent + httpXForwardedFor + requestID + requestTime + \
        upstreamResponseTime + upstreamConnectTime
    # Open access.log (create it if it doesnt exist) and set it to append mode.
    # Using with will autoclose the file, even if there is an exception.
    with open('access.log', 'a') as f:
        # the surrounding single ticks are from nginx.
        f.write(log + "\n")
    f.closed
    sleep(sleepTime)
