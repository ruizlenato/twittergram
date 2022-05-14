import tweepy
from twittergram.config import consumer_key, consumer_secret, access_token, access_token_secret

# Tweepy Things
auth = tweepy.OAuth1UserHandler(consumer_key, consumer_secret, access_token, access_token_secret)
apitweepy = tweepy.API(auth)