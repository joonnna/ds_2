#!/usr/bin/env python
import argparse
import httplib
import uuid


def parse_args():
    parser = argparse.ArgumentParser(prog="client", description="DHT client")

    parser.add_argument("--nameserver", type=str, required=True,
            help="address (host:port) of nameserver to register with")

    return parser.parse_args()


def get_list_of_nodes(nameserver):
    conn = httplib.HTTPConnection(nameserver)
    conn.request("GET", "/", "")

    resp = conn.getresponse()
    addresses = resp.read()

    conn.close()

    return addresses.split()

def generate_pairs(count):
    pairs = {}
    for x in range(0, count):
        key = str(uuid.uuid4())
        value = str(uuid.uuid4())
        pairs[key] = value
    return pairs

def put_value(node, key, value):
    print(node)
    conn = httplib.HTTPConnection(node)
    conn.request("PUT", "/"+key, value)
    conn.getresponse()
    conn.close()

def get_value(node, key):
    conn = httplib.HTTPConnection(node)
    conn.request("GET", "/"+key)
    resp = conn.getresponse()
    value = resp.read().strip()
    conn.close()
    return value

if __name__ == "__main__":

    args = parse_args()

    nodes = get_list_of_nodes(args.nameserver)
    print("%d nodes registered: %s" % (len(nodes), ", ".join(nodes)))

    if len(nodes)==0:
        raise RuntimeError("No nodes registered to connect to")

    pairs = generate_pairs(20)

    tries = 0
    successes = 0
    node_index = 0
    for key, value in pairs.iteritems():
        put_value(nodes[node_index], key, value)
        returned = get_value(nodes[node_index], key)

        tries+=1
        if returned == value:
            successes+=1

        node_index = (node_index+1) % len(nodes)

    success_percent = float(successes) / float(tries) * 100
    print "Stored and retrieved %d pairs of %d (%.1f%%)" % (
            successes, tries, success_percent )
