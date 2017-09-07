'use strict';

window.ee = new EventEmitter();

var OneResult = React.createClass({
    render: function () {
        var data = this.props.data;
        var error = this.props.error;
        var pearsonLists;
        var dataTemplate;
        if (data && data.length > 0) {
            pearsonLists = data.map(function (item, index) {
                return (
                    <tr key={index}>
                        <td>{index + 1}</td>
                        <td><a target="_blank" href={"http://myanimelist.net/profile/"+item.UserName}>{item.UserName}</a></td>
                        <td>{item.Shared}</td>
                        <td>{Math.round(item.Pearson * 100)}</td>
                        <td>{moment(item.LastOnline).isAfter('2004-10-21') ? moment(item.LastOnline).format("ll"): ""}</td>
                        <td>{item.Birthday}</td>
                        <td>{item.Gender}</td>
                        <td>{item.Joined}</td>
                        <td>{item.Location}</td>
                    </tr>
                )
            });
            dataTemplate = (<table className="table table-striped table-hover">
                <thead>
                <tr>
                    <th>#</th>
                    <th>UserName</th>
                    <th>Shared</th>
                    <th>Affinity</th>
                    <th>Last Online</th>
                    <th>Birthday</th>
                    <th>Gender</th>
                    <th>Joined</th>
                    <th>Location</th>
                </tr>

                </thead>
                <tbody>
                {pearsonLists.slice(0, 100)}
                </tbody>
            </table>);
        } else {
            dataTemplate = <p>Empty result</p>
        }
        var errorsTemplate;
        if (error) {
            errorsTemplate = <div className="alert alert-danger" role="alert">{error}</div>
        }

        return (
            <div>
                {errorsTemplate}
                {dataTemplate}
            </div>
        );
    }
});

var CountNewResult = React.createClass({
    getInitialState: function () {
        return {
            isReady: true
        };
    },
    handleKeyPress: function (e) {
        if (e.key === 'Enter') {
            this.onBtnClickHandler(e)
        }
    },
    onBtnClickHandler: function (e) {
        e.preventDefault();
        var share = parseInt(ReactDOM.findDOMNode(this.refs.share).value);
        var anime = ReactDOM.findDOMNode(this.refs.anime).checked;
        var manga = ReactDOM.findDOMNode(this.refs.manga).checked;
        var user = this.props.data;
        if ((anime || manga) && share > 0) {
            this.setState({isReady: false});
            $.post('/api/result/', JSON.stringify({'UserName': user, 'Share': share, 'Anime': anime, 'Manga': manga})).done(function (data) {
                this.setState({isReady: true});
                window.ee.emitEvent('Result.add', [JSON.parse(data)['Data']]);
            }.bind(this)).fail(function() {
                this.setState({isReady: true});
            }.bind(this));
        }
    },
    render: function () {
        var state = this.state;
        return (
            <div className="well well-sm">
                <div className="row">
                    <div className="col-lg-3">
                        <form className="form">
                            <span id="helpBlock" class="help-block">Shared titles</span>
                            <div className="input-group">
                                <input type='text' className='form-control add__user' onKeyPress={this.handleKeyPress}
                                       placeholder='Share count' ref='share' defaultValue='10'/>
                                <span className="input-group-btn">
                                    <button className='add__btn btn btn-default' type="button"
                                            ref='count_button' disabled={!state.isReady}
                                            onClick={this.onBtnClickHandler}> Count
                                    </button>
                                </span>
                            </div>
                            <div className="checkbox">
                                <label>
                                    <input type="checkbox" defaultChecked ref="anime"/>Anime
                                </label>
                            </div>
                            <div className="checkbox">
                                <label>
                                    <input type="checkbox" ref="manga"/>Manga
                                </label>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        );
    }
});

var ResultRow = React.createClass({
    getInitialState: function () {
        return {
            data: {Created: "", Errors: [], Pearson: []},
            hide: false
        };
    },
    componentDidMount: function () {
        this.setState({data: this.props.data});
        this.updateTask(this.props.data.Status);

    },
    getUrl: function() {
        return '/api/result/' + this.props.data.Id;
    },
    updateTask: function(status) {
        status = status || this.state.data.Status;
        if (status == 'counting' || status == 'pending') {
            $.getJSON(this.getUrl()).done(function (data) {
                this.setState({data: data['Data']});
                setTimeout(this.updateTask, 3000)
            }.bind(this))
        }

    },
    onRemoveBtnClickHandler: function (e) {
        e.preventDefault();
        $.ajax({url: this.getUrl(), type: 'DELETE'});
        this.setState({hide: true});
    },
    render: function() {
        var item = this.state.data;
        var hide = this.state.hide;
        var statusClass = 'panel-default';
        if (item.Status == 'complited') {
            statusClass = 'panel-success';
        }
        if (item.Status == 'error') {
            statusClass = 'panel-danger'
        }
        var date = moment(item.Created).format('MMM D LT');
        return (
            <div className={"panel " + statusClass + (hide?" hidden": "")}>
                <div className="panel-heading" role="tab" id="headingOne">
                    <h4 className="panel-title">
                        <a role="button" data-toggle="collapse" data-parent="#accordion"
                           href={'#'+item.Id} className="result__header"
                           aria-expanded="true" aria-controls={item.Id}>
                            <button type="button" className="btn btn-default btn-xs btn-success">
                              <span className="glyphicon glyphicon-arrow-down" aria-hidden="true"></span> Open
                            </button>{' '}
                            <span className="label label-info">{date}</span>{' '}
                            {item.Anime ? <span className="label label-primary">Anime</span>: ''}{' '}
                            {item.Manga ? <span className="label label-primary">Manga</span>: ''}{' '}
                            <span className="label label-primary">Share: {item.Share}</span>{' '}
                            <span className="label label-info">Compare: {item.Compare}</span>{' '}
                            <span className="label label-danger">Errors: {item.Error ? 1:0}</span>{' '}
                            <span className="glyphicon glyphicon-remove result__remove" style={{float: 'right'}}
                                  onClick={this.onRemoveBtnClickHandler} aria-hidden="true"></span>
                            <div className="progress" style={{width: '40%', float: 'right'}}>
                              <div className="progress-bar progress-bar-success progress-bar-striped" role="progressbar"
                                   aria-valuenow={item.Progress} aria-valuemin="0" aria-valuemax="100" style={{width: item.Progress+'%'}}>
                                <span className="sr-only">Progress</span>
                              </div>
                            </div>
                        </a>
                    </h4>
                </div>
                <div id={item.Id} className="panel-collapse collapse" role="tabpanel"
                     aria-labelledby="headingOne">
                    <div className="panel-body">
                        <OneResult data={item.Pearson} error={item.Error} />
                    </div>
                </div>
            </div>
        )
    }
});

var AllResults = React.createClass({
    getInitialState: function () {
        return {
            results: [],
            loading: false,
            first: true,
            user: ""
        };
    },
    componentDidMount: function () {
        var self = this;

        window.ee.addListener('User.add', function (user) {
            self.setState({loading: true, first: false});
            $.getJSON('/api/result/?username=' + user).done(function (data) {
                self.setState({loading: false, results: data['Data'], user: user});
            });
        });

        window.ee.addListener('Result.add', function (result) {
            self.setState({results: [result].concat(self.state.results)});
        });

    },
    render: function () {
        var data = this.state.results;
        var user = this.state.user;
        var state = this.state;
        var resultLists;
        if (data.length > 0) {
            var resultPanelLists = data.map(function (item, index) {
                return <ResultRow key={item.Id} data={item} />
            });
            resultLists = (<div className="panel-group" id="accordion" role="tablist" aria-multiselectable="true">
                {resultPanelLists}
            </div>)
        } else {
            resultLists = <div className="well well-sm">Results not found</div>
        }

        var result;
        if (state.loading) {
            result = <div className="well well-sm">Loading</div>
        } else {
            if (state.first) {
                result = ''
            } else {
                result = <div><CountNewResult data={user}/>{resultLists}</div>
            }
        }
        return (
            <div>
                {result}
            </div>
        );
    }
});

var UserInput = React.createClass({
    getInitialState: function () {
        return {
            userIsEmpty: true
        };
    },
    componentDidMount: function () {
        ReactDOM.findDOMNode(this.refs.user).focus();
    },
    onBtnClickHandler: function (e) {
        e.preventDefault();
        window.ee.emitEvent('User.add', [ReactDOM.findDOMNode(this.refs.user).value]);
    },
    onUserNameChange: function (e) {
        if (e.target.value.trim().length > 0) {
            this.setState({userIsEmpty: false})
        } else {
            this.setState({userIsEmpty: true})
        }
    },
    handleKeyPress: function (e) {
        if (e.key === 'Enter') {
            window.ee.emitEvent('User.add', [ReactDOM.findDOMNode(this.refs.user).value]);
        }
    },
    render: function () {
        var userIsEmpty = this.state.userIsEmpty;
        return (
            <div className='user well well-sm'>
                <div className="row">
                    <div className="col-lg-3">
                        <div className="input-group">
                            <input type='text' className='form-control add__user' onChange={this.onUserNameChange}
                                   placeholder='UserName' ref='user' onKeyPress={this.handleKeyPress}/>
                            <span className="input-group-btn">
                                <button className='add__btn btn btn-default' type="button"
                                        onClick={this.onBtnClickHandler}
                                        ref='user_button' disabled={userIsEmpty}> Set
                                </button>
                            </span>
                        </div>
                    </div>
                    <div className="col-lg-6">
                        <a style={{margin: "20px"}} target="_blank" href="http://linime.animesos.net/">Simple anime game</a>
                        <a style={{margin: "20px"}} target="_blank" href="http://search.animesos.org/">Anime search</a>
                    </div>
                </div>
            </div>
        );
    }
});

var App = React.createClass({
    render: function () {
        return (
            <div className="app">
                <UserInput />
                <AllResults />
            </div>
        );
    }
});

ReactDOM.render(
    <App />,
    document.getElementById('root')
);