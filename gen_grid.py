"""Generate a grid of interleaving mergers and splitters connected by belts with the desired size

Used to test performances.
"""
import sys
import math

def main():
    n = 100
    if len(sys.argv) > 1:
        try:
            n = int(sys.argv[1])
            if n <= 0:
                raise ValueError("n must be positive")
        except (ValueError, TypeError):
            print("Invalid argument, expected a positive integer")

    filename = f"grid_{n}.satisfied"
    with open(filename, "w") as f:
        f.write("#VERSION=0\n")
        for x in range(0, n):
            for y in range(0, n):
                className = "Merger" if (x + y) % 2 else "Splitter"
                f.write(f"{className} {8*x} {8*y} 0\n")
                if x > 0:
                    start = f"{8*(x-1) + 2} {8*y}"
                    end = f"{8*x - 2} {8*y}"
                    if x % 2 == 0:
                        f.write(f"Belt {start} {end}\n")
                    else:
                        f.write(f"Belt {end} {start}\n")
                if y > 0:
                    start = f"{8*x} {8*(y-1) + 2}"
                    end = f"{8*x} {8*y - 2}"
                    if y % 2 == 0:
                        f.write(f"Belt {start} {end}\n")
                    else:
                        f.write(f"Belt {end} {start}\n")
                    
    print("Generated grid:")
    print(f"  - {n} x {n} buildings")
    print(f"  - {math.ceil(n*n/2)} splitters")
    print(f"  - {math.floor(n*n/2)} mergers")
    print(f"  - {(n-1)*(n-1)} belts\n")
    print(filename)

if __name__ == "__main__":
    main()
