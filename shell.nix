
{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.gcc
    pkgs.glibc
    pkgs.pkg-config
    pkgs.docker
  ];
}

