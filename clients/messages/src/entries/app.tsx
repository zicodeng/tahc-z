import * as React from 'react';
import * as ReactDOM from 'react-dom';

import 'stylesheets/entries/app.scss';

class App extends React.Component<any, any> {
    private count: number = 0;

    constructor(props, context) {
        super(props, context);
    }

    public render() {
        return (
            <div>
                <h1>Tahc-Z</h1>
            </div>
        );
    }
}

ReactDOM.render(<App />, document.getElementById('app'));
