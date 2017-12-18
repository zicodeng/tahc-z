import * as React from 'react';

class FloatingActionButton extends React.Component<any, any> {
    constructor(props, context) {
        super(props, context);

        this.state = {
            isFabClicked: false
        };
    }

    public render() {
        return (
            <div className="fab">
                <div
                    className={this.state.isFabClicked ? 'plus active' : 'plus'}
                    onClick={this.handleClickFab}
                >
                    <div className="plus__bar plus__bar--vertical" />
                    <div className="plus__bar plus__bar--horizontal" />
                </div>
                <ul className={this.state.isFabClicked ? 'active' : ''}>
                    {this.props.channels.map((channel, i) => {
                        return (
                            <li key={i} onClick={e => this.handleSelectChannel(e, i)}>
                                {channel.name}
                            </li>
                        );
                    })}
                    <li className="new-channel" onClick={() => this.handleClickNewChannel()}>
                        Create a New Channel
                    </li>
                </ul>
            </div>
        );
    }

    private handleClickFab = (): void => {
        let isFabClicked = this.state.isFabClicked;
        this.setState({
            isFabClicked: !isFabClicked
        });
    };

    private handleSelectChannel = (e, channelIndex): void => {
        this.props.getCurrentChannelIndex(channelIndex);
    };

    private handleClickNewChannel = (): void => {
        this.props.openModal('NewChannel');
    };
}

export default FloatingActionButton;
