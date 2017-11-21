import * as React from 'react';
import axios from 'axios';

class DeleteChannelModal extends React.Component<any, any> {
    constructor(props, context) {
        super(props, context);
        this.state = {};
    }

    public render() {
        return (
            <div className="modal">
                <div className="material-form">
                    <h1 className="title">Delete Channel</h1>
                    <form onSubmit={e => this.handleSubmitForm(e)}>
                        <div className="text-container">
                            <p>Are you sure you want to continue deleting this channel?</p>
                            <p>
                                Deleting this channel will also permanently delete all messages in
                                this channel.
                            </p>
                        </div>
                        <div className="button-container">
                            <button>
                                <span>Confirm</span>
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        );
    }

    private handleSubmitForm = e => {
        e.preventDefault();
        const currentChannel = this.props.currentChannel;
        const host = this.props.host;
        const sessionToken = this.props.sessionToken;
        const url = `https://${host}/v1/channels/${currentChannel._id}`;
        axios
            .delete(url, {
                headers: {
                    Authorization: sessionToken
                }
            })
            .then(res => {
                this.props.closeModal();
            })
            .catch(error => {
                window.alert(error.response.data);
            });
    };
}

export default DeleteChannelModal;
