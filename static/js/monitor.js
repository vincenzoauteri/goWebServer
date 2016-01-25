console.log("Staring Monitor.js");
var url=window.location.href 
console.log(url)
var myPattern="/accounts/([a-zA-Z0-9]+)/venues/([a-zA-Z0-9]+)/stocks/([a-zA-Z0-9]+)$"

var matches= url.match(myPattern);
console.log(matches)

var account= matches[1]
console.log("Account:"+account);
var venue= matches[2]
console.log("Venue:"+venue);
var stock= matches[3]
console.log("Stock:"+stock);

var quotesArray = [];
var format = d3.time.format("%Y-%m-%dT%H:%M:%S");

var margin = {top: 20, right: 20, bottom: 30, left: 50},
    width = 960 - margin.left - margin.right,
    height = 500 - margin.top - margin.bottom;

var formatDate = d3.time.format("%d-%b-%y");

var xTime = d3.time.scale()
    .range([0, width]);

var xIndex = d3.scale.linear()
    .range([0, width]);

var y = d3.scale.linear()
    .range([height,0]);


var xTimeAxis = d3.svg.axis()
    .scale(x)
    .orient("bottom")
    .tickFormat(d3.time.format("%Hh %Mm %Ss"))
    .ticks(d3.time.seconds,5);

var xIndexAxis = d3.svg.axis()
    .scale(xIndex)
    .orient("bottom")
    .tickFormat(function (d,i) {
        if (quotesArray[d] != null ) {
        var qt = quotesArray[d].quoteTime;
        return qt.getHours()+":"+qt.getMinutes()+":"+qt.getSeconds();
        }
    })
    .ticks(5);

var yAxis = d3.svg.axis()
    .scale(y)
    .orient("left")
    .ticks(5);

var x=xIndex;
var xAxis=xIndexAxis;


var chartSvg = d3.select("#chart").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    chartSvg.append("g")
    .attr("class", "x axis")
    .attr("transform", "translate(0," + height + ")");


chartSvg.append("g")
    .attr("class", "y axis")
    .call(yAxis)
    .append("text")
    .attr("transform", "rotate(-90)")
    .attr("y", 6)
    .attr("dy", ".71em")
    .style("text-anchor", "end")
    .text("Price ($)");

var textSvg = d3.select("body").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom);


var line = d3.svg.line()
    .x(function(d) { return x(d.index); })
    .y(function(d) { return y(d.price/100); })
    .interpolate("linear");

var area= d3.svg.area()
    .x(function(d) { return x(d.quoteTime); })
    .y0(height)
    .y1(function(d) { return y(d.price/100); });

chartSvg.append("path")
    .attr("class", "line")
    .attr("stroke", "black")
    .attr("stroke-width", 1)
    .attr("fill", "none");

var quoteCounter = 0;

var quotesWebSocket = new WebSocket("wss://api.stockfighter.io/ob/api/ws/"+account+"/venues/"+venue+"/tickertape/stocks/"+stock);



quotesWebSocket.onmessage = function (event) {
    var ticker = JSON.parse(event.data);
    quoteCounter++;
    //var quote = { "quoteTime":format.parse(ticker.quote.quoteTime.slice(0,19)), "price":ticker.quote.last, "index":quoteCounter};
    var quote = { "quoteTime":new Date(), "price":ticker.quote.last, "index":quoteCounter};

    if (quotesArray.length > 0 && ((new Date(quote.quoteTime)).getTime() > (new Date(quotesArray[quotesArray.length])).getTime())) {
        quotesArray.push(quote);
    } else {
        quotesArray.push(quote);
    }
}

var tick = 0;

var graphQuoteArray = [];
for (var i = 0; i < width; i++) {
        graphQuoteArray.push({ "quoteTime":new Date(), "price":0, "index":-width+i+1});
}


var duration = 0;
var durationToggle = 1;
var counter = 0;
var avg = 0;


function updateGraph() {
    //console.log("Loop cycle: "+(tick++));


    if (quotesArray.length > 0) {

        var lastGraphIndex =  graphQuoteArray[graphQuoteArray.length-1].index;
        var lastQuoteIndex =  quotesArray[quotesArray.length-1].index;

        if (lastGraphIndex < lastQuoteIndex) { 
            graphQuoteArray[graphQuoteArray.length] = quotesArray[lastGraphIndex];


            var xExtent = d3.extent(graphQuoteArray, function(quote){return quote.index;}) 
            x.domain(xExtent);

            var yExtent = d3.extent(graphQuoteArray, function(quote) { return quote.price; });
            y.domain([yExtent[0]/100-(yExtent[0]/100)*0.10,yExtent[1]/100+(yExtent[1]/100)*0.10]);

            chartSvg.select(".line")
            .attr("d",line(graphQuoteArray))
            .transition()
            .duration(duration)
            .attr("transform",null);

            chartSvg.select("g.x.axis").call(xAxis);
            chartSvg.select("g.y.axis").call(yAxis);


            graphQuoteArray.shift();
        } 
    } 
    var interval = 10
    if (lastQuoteIndex - lastGraphIndex< 200) {
        interval = 40;
    } 
    setTimeout(updateGraph, interval);

}
setTimeout(updateGraph, 100);



/*
var accounts = {}

setInterval (function() {
    $.getJSON( "/update/"+venue,  function( data ) {
        for (var i in data) {
        accounts[i] = { account: data[i].Id, NAV:data[i].NAV};
        } 
        console.log(accounts)
    }); 
}, 1000);
*/


var AccountRow = React.createClass({
    render: function() {
        return (
            <tr>
                <td>
                    {this.props.Account}
                </td>
                <td>
                    {this.props.NAV}
                </td>
            </tr>
        );
    }
});

var AccountTable= React.createClass({
    render: function() {
        var accountRows = this.props.data.map(function(position) {
            return (
            <tr key={position.Id}>
                <td>
                    {position.Id}
                </td>
                <td>
                    {position.NAV}
                </td>
                <td>
                    {position.Owned[stock]}
                </td>
                <td>
                    {position.TotBought}
                </td>
                <td>
                    {position.TotSold}
                </td>
            </tr>
            );
        });
        return (
            <table className="account-table">
            <thead>
            <tr>
                <td>
                    Account
                </td>
                <td>
                    NAV
                </td>
                <td>
                    Owned
                </td>
                <td>
                    TotBought
                </td>
                <td>
                    TotSold
                </td>
            </tr>
            </thead>
            <tbody>
            {accountRows}
            </tbody>
            </table>
        );
    }
});

function compareNAV(a,b) {
    if (a.NAV > b.NAV)
        return -1;
    else if (a.NAV < b.NAV)
        return 1;
    else 
        return 0;
}

var AccountsDiv = React.createClass({
    loadDataFromServer: function() {
        $.ajax({
            url: this.props.url,
            dataType: 'json',
            cache: false,
            success: function(data) {
                data.sort(compareNAV);
                this.setState({data: data});
            }.bind(this),
            error: function(xhr, status, err) {
                console.error(this.props.url, status, err.toString());
            }.bind(this)
        });
    },
    getInitialState: function() {
        return {data: []};
    },
    componentDidMount: function() {
        this.loadDataFromServer();
        setInterval(this.loadDataFromServer, this.props.pollInterval);
    },
    render: function() {
        return (
            <div className="accounts-div">
            <h1>Accounts Status</h1>
            <AccountTable data={this.state.data} />
            </div>
        );
    }
});


ReactDOM.render(
    <AccountsDiv url={"/update/"+venue} pollInterval={2000} />,
    document.getElementById("account-table-div")
);
