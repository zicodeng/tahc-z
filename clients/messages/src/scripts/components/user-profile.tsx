import * as React from 'react';

class UserProfile extends React.Component<any, any> {
    constructor(props, context) {
        super(props, context);

        this.state = {
            isEdit: false,
            error: ''
        };
    }

    render() {
        return (
            <div className="profile-card">
                <div className="user-info">
                    <div
                        className="user-photo"
                        style={{ backgroundImage: 'url(' + this.props.user.photoURL + ')' }}
                    />
                    <form className="profile-form" onSubmit={e => this.handleSubmitForm(e)}>
                        <div className="username">
                            <label htmlFor="username">Username</label>
                            <input type="text" value={this.props.user.userName} disabled />
                        </div>
                        <div>
                            <label htmlFor="firstname">First Name</label>
                            <input
                                className={this.state.isEdit ? 'edit' : ''}
                                type="text"
                                ref="firstName"
                                disabled={!this.state.isEdit}
                                required
                                {...(this.state.isEdit
                                    ? null
                                    : { value: this.props.user.firstName })}
                            />
                        </div>
                        <div>
                            <label htmlFor="lastname">Last Name</label>
                            <input
                                className={this.state.isEdit ? 'edit' : ''}
                                type="text"
                                ref="lastName"
                                disabled={!this.state.isEdit}
                                required
                                {...(this.state.isEdit
                                    ? null
                                    : { value: this.props.user.lastName })}
                            />
                        </div>
                        <div>
                            <label htmlFor="email">Email</label>
                            <input type="email" value={this.props.user.email} disabled />
                        </div>
                        <button type="submit">{this.state.isEdit ? 'Save' : 'Edit'}</button>
                    </form>
                    <p className="error">{this.state.error}</p>
                </div>
            </div>
        );
    }

    // Submit form and update user info.
    private handleSubmitForm = (e): void => {
        e.preventDefault();
        let isEdit = this.state.isEdit;
        this.setState({
            isEdit: !isEdit
        });

        const firstName = this.refs.firstName['value'];
        const lastName = this.refs.lastName['value'];

        // Perform submit action only if the form is in edit state
        // and the input values are changed.
        if (
            isEdit &&
            (firstName !== this.props.user.firstName || lastName !== this.props.user.lastName)
        ) {
            const update: Object = {
                firstName: firstName,
                lastName: lastName
            };

            const sessionToken = this.props.getSessionToken();

            let url;
            if (window.location.hostname === 'info-344.zicodeng.me') {
                url = 'https://info-344-api.zicodeng.me/v1/users/me';
            } else {
                url = 'https://localhost/v1/users/me';
            }

            // Validate this session token.
            fetch(url, {
                method: 'PATCH',
                mode: 'cors',
                body: JSON.stringify(update),
                headers: new Headers({
                    Authorization: sessionToken
                })
            })
                .then(res => {
                    if (res.status < 300) {
                        return res.json();
                    }
                    return res.text();
                })
                .then(data => {
                    if (typeof data === 'string') {
                        throw Error(data);
                    } else {
                        this.setState({
                            error: ''
                        });
                        // Pass the updated User to parent component.
                        this.props.user.firstName = data.firstName;
                        this.props.user.lastName = data.lastName;
                        this.props.updateUser(this.props.user);
                    }
                })
                .catch(error => {
                    this.setState({
                        error: error.message
                    });
                });
        }
    };
}

export default UserProfile;
