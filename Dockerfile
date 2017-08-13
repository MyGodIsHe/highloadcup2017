FROM centos:7
WORKDIR /root
ADD traveler traveler
EXPOSE 80
CMD ./traveler
