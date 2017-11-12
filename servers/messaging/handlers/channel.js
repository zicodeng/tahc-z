// @ts-check
'use strict';

const mongodb = require('mongodb');
const express = require('express');
const Channel = require('./../models/channels/channel');
const Message = require('./../models/messages/message');

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
        const message = new Message(channelID, req.body.body, user);
        messageStore
            .insert(message)
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
                // If the current user isn't the creator,
                // respond with the status code 403 (Forbidden).
                if (channel.creator.id !== user.id) {
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
                return channelStore.update(channelID, updates);
            })
            .then(updatedChannel => {
                res.json(updatedChannel);
            })
            .catch(err => {
                throw err;
            });
    });

    // If the current user created the channel, delete it and all messages related to it.
    // If the current user isn't the creator, respond with the status code 403 (Forbidden).
    router.delete('/v1/channels/:channelID', (req, res) => {
        const userJSON = req.get('X-User');
        const user = JSON.parse(userJSON);
        const channelID = new mongodb.ObjectID(req.params.channelID);
        channelStore
            .get(channelID)
            .then(channel => {
                if (channel.creator.id !== user.id) {
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
                throw err;
            });
    });

    return router;
};

module.exports = ChannelHandler;
