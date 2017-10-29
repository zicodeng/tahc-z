import * as React from 'react';
import * as ReactDOM from 'react-dom';

import 'stylesheets/entries/index.scss';

class Index extends React.Component<any, any> {
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

ReactDOM.render(<Index />, document.getElementById('index'));
