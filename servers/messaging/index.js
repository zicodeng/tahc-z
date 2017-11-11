// @ts-check
'user strict';

const mongodb = require('mongodb');
const mongoAddr = process.env.DBADDR || '192.168.99.100:27017';
const mongoURL = `mongodb://${mongoAddr}/info_344`;
const ChannelStore = require('./models/channels/channel-store');
const MessageStore = require('./models/messages/message-store');

const express = require('express');
const app = express();
const morgan = require('morgan');

const addr = process.env.ADDR || 'localhost:4000';
const [host, port] = addr.split(':');
const portNum = parseInt(port);

// Guarantee our MongoDB is started before clients can make any connections.
mongodb.MongoClient
    .connect(mongoURL)
    .then(db => {
        let channelStore = new ChannelStore(db, 'channels');
        let messageStore = new ChannelStore(db, 'messages');

        // Add global middlewares.
        app.use(morgan(process.env.LOG_FORMAT || 'dev'));
        // Parses posted JSON and makes
        // it available from req.body.
        app.use(express.json());

        // All of the following APIs require the user to be authenticated.
        // If the user is not authenticated,
        // respond immediately with the status code 401 (Unauthorized).
        app.use('/', (req, res, next) => {
            const user = req.get('X-User');
            if (!user) {
                res.set('Content-Type', 'text/plain');
                res.status(401).send('no X-User header in the request');
            }
            // Invoke next chained handler if the user is authenticated.
            next();
        });

        // API resource handlers.

        // Error handler.
        app.use((err, req, res, next) => {
            // Write a stack trace to standard out,
            // which writes to the server's log.
            console.error(err.stack);

            // But only report the error message
            // to the client, with a 500 status code.
            res.set('Content-Type', 'text/plain');
            res.status(500).send(err.message);
        });

        app.listen(portNum, host, () => {
            console.log(`server is listening at http://${addr}`);
        });
    })
    .catch(err => {
        throw err;
    });
