import fileinput
import matplotlib.pyplot as plt
import numpy as np

def load_points():
    points = []
    for line in fileinput.input():
        points.append(tuple(map(float, line.split(','))))
    return points

def plot(points):
    sample = points[0]
    points = points[1:]
    fig = plt.figure()
    ax = fig.add_subplot(projection='3d')
    X, Y, Z, Nx, Ny, Nz = zip(*points)
    ax.scatter(X, Y, Z)
    # ax.quiver(*sample, color='black')
    ax.quiver(X, Y, Z, Nx, Ny, Nz, color='black', length=0.2)
    ax.set_xlabel('X')
    ax.set_ylabel('Y')
    ax.set_zlabel('Z')
    plt.show()

def main():
    points = load_points()
    plot(points)

if __name__ == '__main__':
    main()
