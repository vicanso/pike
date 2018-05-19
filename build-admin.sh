#! /bin/bash
cd admin \
  && yarn \
  && yarn build \
  && rm ./dist/js/*.map \
  && cd .. \
  && packr -z
