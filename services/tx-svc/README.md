### Summary

Transaction service is being designed to pull all the bank data and then categorize the spend accordingly. 
This service will be accessible solely via HTTP REST protocol. 

Downstream it relies on the bank service and user service and talks to both of them. 
It is called by the UI which will generate HTTP requests based on user flow.

Behavior: 

By default provides the Txs for a year (if no year is picked by the user)

Future scope,

- extend APIs GET Txs and GET Category to accept account/bank in request
- build functionality to identify copies of the same txs


API Design, 

- Get all transactions for a user within a time frame

    `HTTP GET /trasactions/user/:uid`

    request_params : {                  
        "date_from": datetime,          // optional for the user - defaults to today
        "date_to": datetime,            // optional - defaults to T - 1Yr
        "duration": "string"
    }

    response: {
        "bank_name": string,
        "account_name": string,
        "account_number": string,
        "transactions" : [
            {
                "tx_id":   integer,
                "tx_name": string,
                "tx_date": datetime,
                "type": string/ enum{"credit", "debit"},
                "category" : string,                                // optional
            },
            {
                "tx_id":   integer,
                "tx_name": string,
                "tx_date": datetime,
                "type": string/ enum{"credit", "debit"},
                "category" : string,                                // optional
            },
        ]
    }
    
- Get spend by category for a given user

    `HTTP GET /trasactions/category/user/:uid`

    request_params : {                  
        "date_from": datetime,          // optional for user - otherwise this defaults to today
        "date_to": datetime,            // optional for user - defaults to T - 1Yr
        "duration": "string"            // optional
    }

    response: {
        "categories": [
            {
                "category": string,
                "amount": float64,
                "count": integer,
                "percentage": float64
            },
            {
                ...
            }
        ]
        "total_spending": float64,
        "date_from": datetime,
        "date_to": datetime,
    }


- Update the category of transaction(s)

    `HTTP PATCH /trasactions`

    request: {
        "transaction_ids": []string,
        "updates" : {
            "category" : string,
        }
    }

    response: 200 OK
