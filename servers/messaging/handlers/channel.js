// @ts-check
'use strict';

const mongodb = require('mongodb');
const express = require('express');
const Channel = require('./../models/channels/channel');
const Message = require('./../models/messages/message');

const getUrls = require('get-urls');
const axios = require('axios');

const ChannelHandler = (channelStore, messageStore) => {
    if (!channelStore || !messageStore) {
        throw new Error('no channel and/or message store found');
    }

    const router = express.Router();

    // Respond with the list of all channels.
    router.get('/v1/channels', (req, res) => {
        channelStore
            .getAll()
            .then(channels => {
                res.json(channels);
            })
            .catch(err => {
                throw err;
            });
    });

    // Create a new channel.
    router.post('/v1/channels', (req, res) => {
        const name = req.body.name;
        if (!name) {
            res.set('Content-Type', 'text/plain');
            res.status(400);
            throw new Error('no channel name found in the request');
        }

        let description = '';
        if (req.body.description) {
            description = req.body.description;
        }

        const userJSON = req.get('X-User');
        const user = JSON.parse(userJSON);
        const channel = new Channel(name, description, user);

        channelStore
            .insert(channel)
            .then(channel => {
                res.json(channel);
            })
            .catch(err => {
                throw err;
            });
    });

    // Respond with the latest 50 messages posted to the specified channel.
    router.get('/v1/channels/:channelID', (req, res) => {
        const channelID = new mongodb.ObjectID(req.params.channelID);
        messageStore
            .getAll(channelID)
            .then(messages => {
                res.json(messages);
            })
            .catch(err => {
                throw err;
            });
    });

    // Create a new message in this channel.
    router.post('/v1/channels/:channelID', (req, res) => {
        const userJSON = req.get('X-User');
        const user = JSON.parse(userJSON);
        const channelID = new mongodb.ObjectID(req.params.channelID);
        const messageBody = req.body.body;
        const message = new Message(channelID, messageBody, user);

        // Get any URLs embeded in the message body.
        const URLs = getUrls(messageBody);

        // For each URL, construct an axios.get() promise
        // and push it to promises array
        // which will be consumed by axios.all() as concurrent requests.
        const promises = [];
        if (URLs.size > 0) {
            const summarySvcAddr = 'http://' + (process.env.SUMMARYSVCADDR || 'localhost:5000');
            for (let URL of URLs) {
                let reqURL = summarySvcAddr + '/v1/summary?url=' + URL;
                promises.push(axios.get(reqURL));
            }
        }

        axios
            .all(promises)
            .then(results => {
                results.map(res => {
                    message.summaries.push(res.data);
                });
                return;
            })
            .then(() => {
                return messageStore.insert(message);
            })
            .then(newMessage => {
                res.json(newMessage);
            })
            .catch(err => {
                throw err;
            });
    });

    // Allow channel creator to modify this channel.
    router.patch('/v1/channels/:channelID', (req, res) => {
        const userJSON = req.get('X-User');
        const user = JSON.parse(userJSON);
        const channelID = new mongodb.ObjectID(req.params.channelID);
        channelStore
            .get(channelID)
            .then(channel => {
                if (!channel) {
                    res.set('Content-Type', 'text/plain');
                    res.status(400);
                    throw new Error('no such channel found');
                }
                // If the current user isn't the creator,
                // respond with the status code 403 (Forbidden).
                if (!channel.creator || channel.creator.id !== user.id) {
                    res.set('Content-Type', 'text/plain');
                    res.status(403);
                    throw new Error('only channel creator can modify this channel');
                }
                return;
            })
            .then(() => {
                const updates = {};
                if (req.body.name) {
                    updates.name = req.body.name;
                }
                if (req.body.description) {
                    updates.description = req.body.description;
                }
                updates.editedAt = Date.now();
                return channelStore.update(channelID, updates);
            })
            .then(updatedChannel => {
                res.json(updatedChannel);
            })
            .catch(err => {
                console.log(err.stack);
                res.send(err.message);
            });
    });

    // If the current user created the channel, delete it and all messages related to it.
    // If the current user isn't the creator, respond with the status code 403 (Forbidden).
    router.delete('/v1/channels/:channelID', (req, res, next) => {
        const userJSON = req.get('X-User');
        const user = JSON.parse(userJSON);
        const channelID = new mongodb.ObjectID(req.params.channelID);
        channelStore
            .get(channelID)
            .then(channel => {
                if (!channel) {
                    res.set('Content-Type', 'text/plain');
                    res.status(400);
                    throw new Error('no such channel found');
                }
                if (!channel.creator || channel.creator.id !== user.id) {
                    res.set('Content-Type', 'text/plain');
                    res.status(403);
                    throw new Error('only channel creator can delete this channel');
                }
                return;
            })
            .then(() => {
                return channelStore.delete(channelID);
            })
            .then(() => {
                return messageStore.deleteAll(channelID);
            })
            .then(() => {
                res.set('Content-Type', 'text/plain');
                res.status(200).send('channel deleted');
            })
            .catch(err => {
                console.log(err.stack);
                res.send(err.message);
            });
    });

    return router;
};

module.exports = ChannelHandler;
