#!/bin/bash

#SBATCH --job-name=Hello_OpenMP
#SBATCH --partition=cpu
#SBATCH --output=%j.out
#SBATCH --error=%j.err
#SBATCH -n 80
#SBATCH --ntasks-per-node=40

ulimit -l unlimited
ulimit -s unlimited

module load gcc
gcc -fopenmp omp_hello.c -o omphello.bin

export OMP_NUM_THREADS=60
./omphello.bin