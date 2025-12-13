# If I want to make a single image to run both go and node, combine them here. 
# Probably best to start with a UBI, install with DNF the stuff to run go and npm. 
# Then just use that image. A lot of the specialized containers like nodejs and go are built 
# with nothing but the stuff for those languages and may or may not be in the same spot that 
# you might expect if you had built them with a standard os. 
# like, where is the go bin inside the go container? you'd normally do a COPY and put it somewhere inthe new image. 
FROM registry.redhat.io/ubi10/nodejs-24
