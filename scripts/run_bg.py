import subprocess

with open("output.txt", "w") as f:
    # stdout=f sends the standard output to the file
    # stderr=subprocess.STDOUT merges error messages into the same file
    process = subprocess.Popen(["go", "run", "."], stdout=f, stderr=subprocess.STDOUT)
