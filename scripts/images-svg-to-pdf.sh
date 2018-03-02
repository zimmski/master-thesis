#!/bin/bash

set -exuo pipefail

cd $ROOT_DIR/images

ls *.svg | while read f; do inkscape --without-gui --export-pdf="${f/\.svg/.pdf}" "$f"; done

ls *.pdf | while read f; do pdfcrop $f $f; done
