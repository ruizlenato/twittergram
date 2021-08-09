import tweepy
from twittergram.config import *

# Tweepy Things
auth = tweepy.OAuthHandler(consumer_key, consumer_secret)
auth.set_access_token(access_token, access_token_secret)
apitweepy = tweepy.API(auth)
