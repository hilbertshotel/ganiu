```
ganiu is a trading bot for the kraken exchange

his inputs are entry, stop-loss and take-profit targets

he waits for a limit order to be filled

then he observes the price

if the price drops below the entry target, he places a stop-loss order

if the price goes above the entry target, he places a take-profit order

ganiu closes shop once no more open orders are left

each iteration is 30 seconds long

he also waits for 1 second in between canelation and placement just in case of lag

ganiu is a good bot
```