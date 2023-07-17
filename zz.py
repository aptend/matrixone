import sys

args = sys.argv

filename = args[1]
base = int(args[2])
gap = int(args[3])

with open(filename, 'w', encoding='utf8') as f:
    for i in range(base, base+gap):
        f.write(f"{i},7,42,\n")
