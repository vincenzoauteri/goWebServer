LEVEL 6 solution:

This code try to solve level 6 of stockfighter (and was used to pass the other levels as well).

The backend in go provide the interface with the SF api and the javascript frontend visualizes stock quotes and update account status in real time.

Account numbers were found by exploiting the response of the "Cancel Order" error response, after trying all the API commands and checking the response.

An buy order is made, and all the orders with numbers inferior to that number are canceled, the error response is parset to get the account string and this is stored. 
Then another order is made etc...

Once the accounts are collected, executions can be monitored by opening a websocket for each account, since the url only needs the account number (as hinted by the documentation).

The position for all accounts is updated according to the fills received and the stock last quote. A list sorted by NAV is displayed on the fronted.

Finally to identify the culprit, a few assumptions: must have a fairly positive NAV, must not trade at a very high frequency, not worry too much about exposure. 

After running a few runs, and analyzing the final summary for each account there is one that seems to be different from the other bots, as having a nice NAV and a relatively low number of orders, compared to the other accounts.

![alt tag](https://raw.github.com/username/projectname/branch/path/to/img.png)
