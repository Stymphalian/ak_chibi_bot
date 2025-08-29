"""
Write a program which reads through every file under a directory and
make a sorted index of every file under that directory. The output should be a json file.

Now take two of these json files and diff them.
List the entries which are in A but not B
and the entries which are in B and not A
Also output those diffs as json
"""

import argparse
import json
import os
import sys

def build_index(directory):
    """Given a directory, return a sorted index of every file under that directory"""
    index = {}
    for root, dirs, files in os.walk(directory):
        for file in files:
            path = os.path.join(root, file)
            index[path] = None
    return dict(sorted(index.items()))

def diff_indexes(indexA, indexB):
    """Given two indexes, return two dictionaries: one with entries in A but not B and one with entries in B but not A"""
    inAnotB = {}
    inBnotA = {}
    for path in indexA:
        if path not in indexB:
            inAnotB[path] = None
    for path in indexB:
        if path not in indexA:
            inBnotA[path] = None
    return inAnotB, inBnotA

def main():
    parser = argparse.ArgumentParser(description='Build or diff two json indexes')
    parser.add_argument('--build', action='store_true', help='build a json index')
    parser.add_argument('--diff', action='store_true', help='diff two json indexes')
    parser.add_argument('--directory', help='directory to build index from')
    parser.add_argument('--indexA', help='first json index')
    parser.add_argument('--indexB', help='second json index')
    args = parser.parse_args()

    if args.build:
        index = build_index(args.directory)
        json.dump(index, sys.stdout, indent=2)
        sys.stdout.write('\n')
    elif args.diff:
        with open(args.indexA) as f:
            indexA = json.load(f)
        with open(args.indexB) as f:
            indexB = json.load(f)
        inAnotB, inBnotA = diff_indexes(indexA, indexB)
        json.dump({'inAnotB': inAnotB, 'inBnotA': inBnotA}, sys.stdout, indent=2)
        sys.stdout.write('\n')
    else:
        parser.print_help()

if __name__ == '__main__':
    main()
