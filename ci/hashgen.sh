#!/bin/sh

for f in bin/inlets*; do shasum -a 256 $f > $f.sha256; done
