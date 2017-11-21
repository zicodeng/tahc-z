import * as React from 'react';
import axios from 'axios';

class Chat extends React.Component<any, any> {
    private MESSAGE_MODE = {
        EDIT: 0,
        CREATE: 1
    };
    constructor(props, context) {
        super(props, context);

        this.state = {
            keys: {},
            textarea: null,
            messageMode: this.MESSAGE_MODE.CREATE,
            selectedMessage: null
        };
    }

    public render(): JSX.Element {
        const channel = this.props.channel;
        return (
            <div className="chat">
                <header className="channel">
                    <div className="channel__info">
                        {channel.description ? <p>{channel.description}</p> : null}
                        {channel.creator ? (
                            <p>
                                {`Created by ${channel.creator.firstName} ${channel.creator
                                    .lastName}`}
                            </p>
                        ) : null}
                    </div>
                    <h3 className="channel__name">{channel.name}</h3>
                    {this.renderChannelActions()}
                </header>
                <div id="message-container" className="message-container">
                    {this.renderMessages()}
                </div>
                <textarea
                    id="textarea"
                    rows={1}
                    ref="messageBody"
                    placeholder={`SHIFT+CTRL to send a message @${channel.name}`}
                    autoFocus
                    onKeyDown={e => this.handleSubmitMessage(e)}
                    onKeyUp={e => this.handleSubmitMessage(e)}
                />
            </div>
        );
    }

    componentDidUpdate() {
        // Always scroll to the bottom .
        const messageContainer = document.getElementById('message-container');
        if (messageContainer) {
            messageContainer.scrollTop = messageContainer.scrollHeight;
        }
    }

    componentDidMount() {
        const textarea = document.getElementById('textarea');
        this.setState({
            textarea: textarea
        });
    }

    private renderChannelActions = (): JSX.Element | null => {
        const creator = this.props.channel.creator;
        const user = this.props.user;
        if (!creator || user.id !== creator.id) {
            return null;
        }
        return (
            <div className="channel__actions">
                <div className="channel__actions--edit" onClick={() => this.handleEditChannel()}>
                    <i className="fa fa-pencil" aria-hidden="true" />
                </div>
                <div
                    className="channel__actions--delete"
                    onClick={() => this.handleDeleteChannel()}
                >
                    <i className="fa fa-trash" aria-hidden="true" />
                </div>
            </div>
        );
    };

    private handleEditChannel = (): void => {
        this.props.openModal('EditChannel');
    };

    private handleDeleteChannel = (): void => {
        this.props.openModal('DeleteChannel');
    };

    private renderMessages = (): JSX.Element => {
        const user = this.props.user;
        const li = this.props.messages.map((msg, i) => {
            return (
                <li key={i} className={msg.creator.id === user.id ? 'editable' : ''}>
                    <div className="message">
                        <div
                            className="photo"
                            style={{ backgroundImage: 'url(' + msg.creator.photoURL + ')' }}
                        />
                        <div className="content">
                            <h4>{`${msg.creator.firstName} ${msg.creator.lastName}`}</h4>
                            <p>{msg.body}</p>
                        </div>
                        {this.renderMessageActions(i)}
                    </div>
                    {this.renderSummaries(msg)}
                </li>
            );
        });
        return <ul>{li}</ul>;
    };

    private renderSummaries = (msg): JSX.Element | null => {
        console.log(msg);
        if (!msg.summaries.length) {
            return null;
        }
        const summaries = msg.summaries.map((summary, i) => {
            return (
                <li className="summary">
                    {summary.images && summary.images[0] ? (
                        <div
                            className="summary__image"
                            style={{ backgroundImage: 'url(' + summary.images[0].url + ')' }}
                        />
                    ) : null}
                    {summary.title ? <h4>{summary.title}</h4> : null}
                    {summary.description ? <p>{summary.description}</p> : null}
                    {summary.url ? (
                        <a href={summary.url} target="_blank" className="summary__link">
                            {summary.url}
                        </a>
                    ) : null}
                </li>
            );
        });
        return <ul className="summaries">{summaries}</ul>;
    };

    private renderMessageActions = (i: number): JSX.Element => {
        return (
            <div className="message__actions">
                <div className="message__actions--edit" onClick={e => this.handleEditMessage(e, i)}>
                    <i className="fa fa-pencil" aria-hidden="true" />
                </div>
                <div
                    className="message__actions--delete"
                    onClick={e => this.handleDeleteMessage(e, i)}
                >
                    <i className="fa fa-trash" aria-hidden="true" />
                </div>
                <div className="divider" />
            </div>
        );
    };

    private handleEditMessage = (e, i): void => {
        // Dummy way to get message content by traversing DOM tree.
        // const oldMessage = e.currentTarget.parentElement.previousSibling.getElementsByTagName(
        //     'p'
        // )[0].innerText;

        const messageToBeEdited = this.props.messages[i];

        const textarea = this.state.textarea;
        textarea.value = messageToBeEdited.body;
        textarea.focus();

        // Set message mode to edit.
        this.setState({
            messageMode: this.MESSAGE_MODE.EDIT,
            selectedMessage: messageToBeEdited
        });
    };

    private handleDeleteMessage = (e, i): void => {
        const messageToBeDeleted = this.props.messages[i];
        const host = this.props.host;
        const sessionToken = this.props.sessionToken;
        const url = `https://${host}/v1/messages/${messageToBeDeleted._id}`;
        axios
            .delete(url, {
                headers: {
                    Authorization: sessionToken
                }
            })
            .catch(error => {
                window.alert(error.response.data);
            });
    };

    private handleSubmitMessage = (e): void => {
        const messageBody: String = this.refs.messageBody['value'];
        // Keeps track of what keys the user is pressing down.
        let keys = this.state.keys;
        // If the key is pressed down, set its value to true in the map.
        // If the key is lifted up, set its value to false in the map.
        keys[e.keyCode] = e.type === 'keydown';

        // If SHIFT + ENTER key is pressed,
        // send this message.
        if (keys[13] && keys[16] && messageBody) {
            e.preventDefault();
            const messageMode = this.state.messageMode;
            const message = {
                body: messageBody
            };
            if (messageMode === this.MESSAGE_MODE.CREATE) {
                this.createMessage(message);
            } else {
                this.editMessage(message);
            }
        }
    };

    private createMessage = (message): void => {
        const channel = this.props.channel;
        const host = this.props.host;
        const sessionToken = this.props.sessionToken;
        const url = `https://${host}/v1/channels/${channel._id}`;
        axios
            .post(url, message, {
                headers: {
                    Authorization: sessionToken
                }
            })
            .then(res => {
                // Clear textarea after the message is created successfully.
                this.refs.messageBody['value'] = '';
            })
            .catch(error => {
                window.alert(error.response.data);
            });
    };

    private editMessage = (message): void => {
        const selectedMessage = this.state.selectedMessage;
        const host = this.props.host;
        const sessionToken = this.props.sessionToken;
        const url = `https://${host}/v1/messages/${selectedMessage._id}`;
        axios
            .patch(url, message, {
                headers: {
                    Authorization: sessionToken
                }
            })
            .then(res => {
                // Clear textarea after the message is edited successfully.
                this.refs.messageBody['value'] = '';
                this.setState({
                    messageMode: this.MESSAGE_MODE.CREATE
                });
            })
            .catch(error => {
                window.alert(error.response.data);
            });
    };
}

export default Chat;
