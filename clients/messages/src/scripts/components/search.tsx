import * as React from 'react';

interface User {
    id: string;
    userName: string;
    firstName: string;
    lastName: string;
    email: string;
    photoURL: string;
}

class Search extends React.Component<any, any> {
    constructor(props, context) {
        super(props, context);

        this.state = {
            query: '',
            suggestions: []
        };
    }

    render() {
        return (
            <div className="search-container">
                <h3>Search Users</h3>
                <input
                    type="text"
                    ref="query"
                    placeholder="Search by first name, last name, username, or email"
                    onChange={e => this.handleChangeInput()}
                />
                {this.renderSuggestions()}
            </div>
        );
    }

    private handleChangeInput = (): void => {
        let suggestions: User[] = [];
        const query = this.refs.query['value'].trim().toLowerCase();
        if (query.length === 0) {
            this.setState({
                suggestions: suggestions
            });
            return;
        }

        let url;
        if (window.location.hostname === 'info-344.zicodeng.me') {
            url = 'https://info-344-api.zicodeng.me/v1/users?q=' + query;
        } else {
            url = 'https://localhost/v1/users?q=' + query;
        }

        fetch(url, {
            method: 'get',
            mode: 'cors',
            headers: new Headers({
                Authorization: this.props.sessionToken
            })
        })
            .then(res => {
                if (!res.ok) {
                    return res.text();
                }
                return res.json();
            })
            .then(data => {
                if (typeof data === 'string') {
                    throw Error(data);
                }
                this.setState({
                    query: query,
                    suggestions: data
                });
            })
            .catch(error => {
                console.log(error);
            });
    };

    private renderSuggestions = (): JSX.Element => {
        const suggestions: User[] = this.state.suggestions;
        let list = suggestions.map((user, i) => {
            return (
                <li key={i}>
                    <div style={{ backgroundImage: 'url(' + user.photoURL + ')' }} />
                    <p>
                        {this.highlightSearch(user.firstName)}
                        <span> </span>
                        {this.highlightSearch(user.lastName)}
                        <span> | </span>
                        {this.highlightSearch(user.userName)}
                        <span> | </span>
                        {this.highlightSearch(user.email)}
                    </p>
                </li>
            );
        });
        return <ul className="suggestion-list">{list}</ul>;
    };

    // Highlight the search query in the displayed suggestions.
    private highlightSearch = (text: string): JSX.Element => {
        let query = this.state.query;
        const tokens: string[] = query.split(' ');
        for (const token of tokens) {
            if (text.toLowerCase().startsWith(token)) {
                // Ensure the result is displayed with original case.
                const highlightToken = text.substring(0, token.length);
                const rest = text.substring(token.length, text.length);
                return (
                    <span>
                        <span className="highlight">{highlightToken}</span>
                        <span>{rest}</span>
                    </span>
                );
            }
        }
        return <span>{text}</span>;
    };
}

export default Search;
