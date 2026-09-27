# Tech case: transactions routine

## Overview

Step 1 is a simple service with 3 endpoints. Step 2 is a live session (up to 1h30,
screen shared) where new features are built together on top of this service.

### Technical notes
- Preferred languages: Java, Groovy, Kotlin or Go.
- The solution must be published on GitHub with a README explaining how to run it.

### Rating criteria
1. Maintainability
2. Simplicity
3. Testability
4. Documentation

### Bonus
- Docker
- Easy execution (`./run` script)
- Good documentation
- Tests

## Domain

Each cardholder (customer) has an account with their data.
For each operation done by the customer, a transaction is created and associated
with their account.

Each transaction has a type: normal purchase, purchase with installments,
withdrawal or credit voucher.

Purchase and withdrawal transactions are registered with **negative** amounts.
Credit voucher transactions are registered with **positive** amounts.

## Suggested data structure

The model may be changed.

### Accounts

| Account_ID | Document_Number |
|------------|-----------------|
| 1          | 12345678900     |

### OperationsTypes

| OperationType_ID | Description                |
|------------------|----------------------------|
| 1                | Normal Purchase            |
| 2                | Purchase with installments |
| 3                | Withdrawal                 |
| 4                | Credit Voucher             |

### Transactions

| Transaction_ID | Account_ID | OperationType_ID | Amount | EventDate                   |
|----------------|------------|------------------|--------|-----------------------------|
| 1              | 1          | 1                | -50.0  | 2020-01-01T10:32:07.7199222 |
| 2              | 1          | 1                | -23.5  | 2020-01-01T10:48:12.2135875 |
| 3              | 1          | 1                | -18.7  | 2020-01-02T19:01:23.1458543 |
| 4              | 1          | 4                | 60.0   | 2020-01-05T09:34:18.5893223 |

`Amount` holds the transaction value. `EventDate` holds the moment the transaction occurred.

## Endpoints

### POST /accounts
Creates an account.

```json
{ "document_number": "12345678900" }
```

### GET /accounts/:accountId
Retrieves the account information.

```json
{ "account_id": 1, "document_number": "12345678900" }
```

### POST /transactions
Creates a transaction.

```json
{ "account_id": 1, "operation_type_id": 4, "amount": 123.45 }
```