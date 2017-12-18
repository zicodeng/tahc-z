import * as React from 'react';
import axios from 'axios';

class EditChannelModal extends React.Component<any, any> {
    constructor(props, context) {
        super(props, context);
        this.state = {};
    }

    public render() {
        const currentChannel = this.props.currentChannel;
        return (
            <div className="modal">
                <div className="material-form">
                    <h1 className="title">Edit Channel</h1>
                    <form onSubmit={e => this.handleSubmitForm(e)}>
                        <div className="input-container">
                            <input
                                type="text"
                                ref="name"
                                defaultValue={currentChannel.name}
                                required
                            />
                            <label htmlFor="name">Name</label>
                            <div className="bar" />
                        </div>
                        <div className="input-container">
                            <input
                                type="text"
                                ref="desc"
                                defaultValue={currentChannel.description}
                            />
                            <label htmlFor="desc">Description</label>
                            <div className="bar" />
                        </div>
                        <div className="button-container">
                            <button>
                                <span>Save</span>
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
        const name = this.refs.name['value'];
        const desc = this.refs.desc['value'];
        const updatedChannel = {
            name: name,
            description: desc
        };
        const host = this.props.host;
        const sessionToken = this.props.sessionToken;
        const url = `https://${host}/v1/channels/${currentChannel._id}`;
        axios
            .patch(url, updatedChannel, {
                headers: {
                    Authorization: sessionToken
                }
            })
            .then(res => {
                // Clear textarea after the message is created successfully.
                this.refs.name['value'] = '';
                this.refs.desc['value'] = '';
                this.props.closeModal();
            })
            .catch(error => {
                window.alert(error.response.data);
            });
    };
}

export default EditChannelModal;
