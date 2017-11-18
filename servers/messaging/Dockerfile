FROM node:alpine
WORKDIR /messaging
COPY . .
RUN npm install --only=prod
EXPOSE 80
ENTRYPOINT ["node", "index.js"]