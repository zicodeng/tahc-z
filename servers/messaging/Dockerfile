FROM node:alpine
WORKDIR /messaging
COPY . .
RUN npm install
EXPOSE 80
ENTRYPOINT ["node", "index.js"]