import json
import time
from imdb import Cinemagoer

FILE = "movies.json"

def read():
	with open(FILE, "r") as f:
		return json.load(f)

def write(info):
	movies = read()
	movies.append(info)
	with open(FILE, "w") as f:
		json.dump(movies, f, indent=4)

def get_info(movie):
	poster = movie["cover url"]
	plot = movie["plot"][0]

	info = {
		"id": movie.movieID,
		"title": movie["title"],
		"year": movie["year"],
		"directors": [d["name"] for d in movie["directors"]],
		"poster": poster[:poster.rfind("._V1")],
		"plot": plot[:plot.rfind("::")],  # removes author info
	}

	write(info)

imdb = Cinemagoer()
top = imdb.get_top250_movies()

movies = read()

for i, movie in enumerate(top[len(top)-1:]):
	movie = imdb.get_movie(movie.movieID)
	print(f"{i+1}/{len(top)}")
	time.sleep(5)
