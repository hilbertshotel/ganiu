```
ganiu is a trading bot for the kraken exchange

his inputs are entry, stop-loss and take-profit targets

he waits for my limit order to be filled, then he observes

if the price goes above the entry target he cancels my stop-loss and places a take-profit

if the price drops below the entry target, he cancels my take-profit and places a stop loss

he does so until one of the orders gets filled

other inputs include wait time and currency pair
```