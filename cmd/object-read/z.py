def sort_lines_by_ts(filename):
    with open(filename, 'r') as file:
        lines = file.readlines()

    # Split each line into components and convert the ts value to an integer
    lines = [line.split(", ") for line in lines]
    for line in lines:
        line[1] = line[1].split(" ")
        line[1][1] = int(line[1][1].split("-")[0])

    # Sort the lines by the ts value
    lines = sorted(lines, key=lambda line: line[1][1])

    # Join the components back into a single string
    lines = [", ".join([part[0] + " " + str(part[1]) if isinstance(part, list) else part for part in line]) for line in lines]

    with open(filename, 'w') as file:
        file.write("\n".join(lines))

sort_lines_by_ts('lines')
