import fileinput
import matplotlib.pyplot as plt
import numpy as np

def load_points():
    points = []
    for line in fileinput.input():
        points.append(tuple(map(float, line.split(','))))
    return points

def dist(a, b):
    return np.linalg.norm(np.array(a) - np.array(b))

def plot(points):
    # sample = points[0]
    # points = points[1:]
    # points = [x for x in points if dist(x, sample) <= 2]
    fig = plt.figure()
    ax = fig.add_subplot(projection='3d')
    # X, Y, Z, Nx, Ny, Nz = zip(*points)
    X, Y, Z = zip(*points)
    ax.plot(X, Y, Z, '-o')
    # ax.scatter(X, Y, Z)
    # ax.quiver(*sample, color='black')
    # ax.quiver(X, Y, Z, Nx, Ny, Nz, color='black', length=0.2)
    ax.set_xlabel('X')
    ax.set_ylabel('Y')
    ax.set_zlabel('Z')
    ax.set_box_aspect((np.ptp(X), np.ptp(Y), np.ptp(Z)))
    plt.show()

def main():
    points = load_points()
    plot(points)

if __name__ == '__main__':
    main()
