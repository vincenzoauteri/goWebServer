console.log("Staring Monitor.js");
var url=window.location.href 
console.log(url)
var myPattern="/accounts/([a-zA-Z0-9]+)/venues/([a-zA-Z0-9]+)/stocks/([a-zA-Z0-9]+)$"

var matches= url.match(myPattern);
console.log(matches)

account= matches[1]
console.log("Account:"+account);
venue= matches[2]
console.log("Venue:"+venue);
stock= matches[3]
console.log("Stock:"+stock);

var quotesArray = [];
var quotes= [];
var format = d3.time.format("%Y-%m-%dT%H:%M:%S");

var margin = {top: 20, right: 20, bottom: 30, left: 50},
    width = 960 - margin.left - margin.right,
    height = 500 - margin.top - margin.bottom;

var formatDate = d3.time.format("%d-%b-%y");

var x = d3.scale.linear()
    .range([0, width]);

var y = d3.scale.linear()
    .range([height, 0]);

var xAxis = d3.svg.axis()
    .scale(x)
    .orient("bottom")
    .ticks(5);

var yAxis = d3.svg.axis()
    .scale(y)
    .orient("left")
    .ticks(5);


var svg = d3.select("body").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    x.domain(d3.extent(quotesArray, function(quote) { return quote.index; })); 
                       //{return  format.parse(quote.quoteTime.slice(0,19))}));
    svg.append("g")
    .attr("class", "x axis")
    .attr("transform", "translate(0," + height + ")")
    .call(xAxis);

    y.domain(d3.extent(quotesArray, function(quote) { return quote.last; }));

    svg.append("g")
    .attr("class", "y axis")
    .call(yAxis)
    .append("text")
    .attr("transform", "rotate(-90)")
    .attr("y", 6)
    .attr("dy", ".71em")
    .style("text-anchor", "end")
    .text("Price ($)");



var quoteCounter = 0;
var quotesWebSocker = new WebSocket("wss://api.stockfighter.io/ob/api/ws/"+account+"/venues/"+venue+"/tickertape/stocks/"+stock);
quotesWebSocker.onmessage = function (event) {
    var ticker = JSON.parse(event.data);

    quoteCounter++;

    var quote = { "quoteTime":ticker.quote.quoteTime, "price":ticker.quote.last, "index":quoteCounter};
    quotesArray[quotesArray.length] = quote;
    quotes[quotes.length] = ticker.quote.last;      

    x.domain(d3.extent(quotesArray, function(quote) { return quote.index; })); 
                       //{return  format.parse(quote.quoteTime.slice(0,19))}));

    y.domain(d3.extent(quotesArray, function(quote) { return quote.last; }));



    
    /*
    var bar = svg.selectAll("g")
    .data(quotesArray)
    .enter().append("g")
    .attr("transform", function(d, i) { return "translate(" + i * 1+ ",0)"; });

    bar.append("rect")
    .attr("height", function(d,i) {return d.price/100;})
    .attr("width", 1);  
    */

    var line= d3.svg.line()
    .x(function(d) { return d.index; })
    .y(function(d) { return d.price/100; })
    .interpolate("linear");

    svg.append("path")
    .attr("d", line(quotesArray))
    .attr("stroke", "blue")
    .attr("stroke-width", 2)
    .attr("fill", "none");

}


/*
//var n = 200;
var random = d3.random.normal(0, .2);

function chart(domain, interpolation, tick) {
    //var data = d3.range(n).map(random);
    var data = quotesArray
    var margin = {top: 6, right: 0, bottom: 6, left: 40};
    var width = 960 - margin.right;
    var height = 960 - margin.top - margin.bottom;

    var x = d3.scale.linear()
    .domain(domain)
    .range([0, width]);

    var y = d3.scale.linear()
    .domain([0, 300])
    .range([height, 0]);

    var line = d3.svg.line()
    .interpolate(interpolation)
    .x(function(d, i) { return x(i); })
    .y(function(d, i) { return y(d); });

    var svg = d3.select("body").append("p").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .style("margin-left", -margin.left + "px")
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    svg.append("defs").append("clipPath")
    .attr("id", "clip")
    .append("rect")
    .attr("width", width)
    .attr("height", height);

    svg.append("g")
    .attr("class", "y axis")
    .call(d3.svg.axis().scale(y).ticks(5).orient("left"));

    var path = svg.append("g")
    .attr("clip-path", "url(#clip)")
    .append("path")
    .datum(data)
    .attr("class", "line")
    .attr("d", line);

    tick(path, line, data, x);
};

var transition = d3.select({}).transition().duration(750).ease("linear");

chart([0, quotesArray.length, "linear", function tick(path, line, data) {
    transition = transition.each(function() {

        // push a new data point onto the back
        data.push(random());
        //
        //         // pop the old data point off the front
        data.shift();
        //
        //                 // transition the line
        path.transition().attr("d",
                               line);

    }).transition().each("start",
    function() { tick(path,
                      line, data); });
});
*/
