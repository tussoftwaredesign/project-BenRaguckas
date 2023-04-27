from consumer import Consumer
from custom_fufnctions import *
import sys


if __name__ == '__main__':
    # API_HOST = "localhost:8080"
    # RMQ_HOST = "localhost"
    API_HOST = "api.bean-bon.eu"
    RMQ_HOST = "143.42.111.216"
    RMQ_PORT = 5672
    options = [
        ('"backgroundQ" for removing image backgrounds.', "backgroundQ", removebackground),
        ('"grayoutQ" for gray scaling images.', "grayoutQ", grayscale),
        ('"videoaudioextractQ" extracts audio track from video file.', "videoaudioextractQ", extractaudiotrack),
        ('"transcribeQ" transcribes audio file.', "transcribeQ", speechrecognition),
    ]
    selected_description, selected_que, selected_function = options[int(sys.argv[1])]
    print(f"Selected option: {selected_description}")
    # if selected_que == "transcribeQ":
    #
    # else:
    consumer = Consumer(selected_que, API_HOST, RMQ_HOST, RMQ_PORT, selected_function)
    consumer.begin()
