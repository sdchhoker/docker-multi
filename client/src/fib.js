import React, { Component } from "react";
import axios from "axios";

class Fib extends Component {
    state = {
        seenIndexes: [],
        values: {},
        index: ''
    }
    componentDidMount() {
        this.fetchValues();
        this.fetchIndexes();
    }
    async fetchValues() {
        const values = await axios.get("/api/values/current");
        console.log(values)
        this.setState({values: values.data ||{}});
    }
    async fetchIndexes() {
        const setIndexes = await axios.get("api/values/all");
        console.log(setIndexes)
        this.setState({seenIndexes: setIndexes.data ||[]});
    }

    handleSubmit = async (event) => {
        event.preventDefault();
        await axios({
            method: "post",
            url: "api/values",
            data: {
                index: parseInt(this.state.index)
            }
        });
        this.setState({ index : ''})
        this.fetchValues();
        this.fetchIndexes();
    }

    render() {
        return (
            <div>
                <form onSubmit={this.handleSubmit}>
                    <input value={this.state.index} onChange={(event) => {
                        this.setState({index: event.target.value})
                    }} />
                    <button type="submit">Submit</button>
                </form>
                <h3>Seen indexes are: </h3>
                <p>{this.state.seenIndexes.join(', ')}</p>
                <h3>Indexes with values are: </h3>
                <div>
                    {Object.keys(this.state.values).map((key) => {
                        return (<p key={key}>
                            {`${key} has value ${this.state.values[key]}`}
                        </p>);
                    })}
                </div>
            </div>
        );
    }
}

export default Fib;
