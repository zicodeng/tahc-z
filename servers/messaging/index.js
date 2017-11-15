// @ts-check
'use strict';

const mongodb = require('mongodb');
const mongoAddr = process.env.DBADDR || '192.168.99.100:27017';
const mongoURL = `mongodb://${mongoAddr}/info_344`;
const ChannelStore = require('./models/channels/channel-store');
const MessageStore = require('./models/messages/message-store');

const redis = require('redis');
const redisAddr = process.env.REDISADDR || '192.168.99.100';

const express = require('express');
const app = express();
const morgan = require('morgan');

const Channel = require('./models/channels/channel');

const ChannelHandler = require('./handlers/channel');
const MessageHandler = require('./handlers/message');

const addr = process.env.ADDR || 'localhost:4000';
const [host, port] = addr.split(':');
const portNum = parseInt(port);

// Guarantee our MongoDB is started before clients can make any connections.
mongodb.MongoClient
    .connect(mongoURL)
    .then(db => {
        // When messaging microservice starts up,
        // publish information about this microservice every 10 seconds,
        // so that our gateway is guaranteed to get the lastest status
        // about this microservice. If it dies, our gateway will be informed.
        const publisher = redis.createClient({
            host: redisAddr
        });
        const sec = 1000;
        const heartBeat = 10;
        const msgSvc = {
            name: 'messaging',
            pathPattern: '/v1/(channels|messages)/?',
            address: addr,
            heartbeat: heartBeat
        };
        setInterval(() => {
            publisher.publish('microservices', JSON.stringify(msgSvc));
        }, 1000 * heartBeat);

        // Add global middlewares.
        app.use(morgan(process.env.LOG_FORMAT || 'dev'));
        // Parses posted JSON and makes
        // it available from req.body.
        app.use(express.json());

        // All of the following APIs require the user to be authenticated.
        // If the user is not authenticated,
        // respond immediately with the status code 401 (Unauthorized).
        app.use((req, res, next) => {
            const userJSON = req.get('X-User');
            if (!userJSON) {
                res.status(401);
                throw new Error('no X-User header found in the request');
            }
            // Invoke next chained handler if the user is authenticated.
            next();
        });

        // Initialize Mongo stores.
        let channelStore = new ChannelStore(db, 'channels');
        let messageStore = new MessageStore(db, 'messages');

        // Add a default channel named "general".
        const channel = new Channel('general', '');
        channelStore.insert(channel);

        // API resource handlers.
        app.use(ChannelHandler(channelStore, messageStore));
        app.use(MessageHandler(messageStore));

        // Error handler.
        app.use((err, req, res, next) => {
            // Write a stack trace to standard out,
            // which writes to the server's log.
            console.error(err.stack);

            // But only report the error message to the client.
            res.set('Content-Type', 'text/plain');
            if (res.statusCode === 200) {
                res.status(500);
            }
            res.send(err.message);
        });

        app.listen(portNum, host, () => {
            console.log(`server is listening at http://${addr}`);
        });
    })
    .catch(err => {
        throw err;
    });
