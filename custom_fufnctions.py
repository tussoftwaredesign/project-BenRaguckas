import sys

import PIL
from io import BytesIO
import subprocess
import whisper

try:
    from rembg import remove
except ImportError:
    print("Could not import rembg")

def removebackground(file, params):
    return "result.png", remove(file)


def grayscale(file, params):
    image = PIL.Image.open(BytesIO(file))
    gray = image.convert("L")
    buffer = BytesIO()
    gray.save(buffer, format="png")
    buffer.seek(0)
    return "result.png", buffer


def extractaudiotrack(file, params):
    video_buffer = BytesIO(file)
    video_buffer.seek(0)
    file_v = "./temp.mp4"
    file_a = "./temp.mp3"
    with open(file_v, 'wb') as temp_vid:
        temp_vid.write(file)
    ffmpeg_command = ['ffmpeg', '-y', '-i', file_v, '-vn', '-acodec', 'libmp3lame', file_a]
    # ffmpeg_command = ['ffmpeg', '-y', '-i', file_v, '-vn', '-acodec', 'pcm_s16le','-ar', '44100', '-ac', '2', file_a]
    subprocess.run(ffmpeg_command, check=True)
    with open(file_a, 'rb') as temp_aud:
        buffer = BytesIO(temp_aud.read())
        buffer.seek(0)
        return "track.wav", buffer


def speechrecognition(file, params):
    video_buffer = BytesIO(file)
    video_buffer.seek(0)
    file_v = "./transcribe.mp3"
    with open(file_v, 'wb') as temp_vid:
        temp_vid.write(file)
    if params is not None and "model" in params:
        model = params['model']
    else:
        model = "base"
    model = whisper.load_model(model)
    audio = whisper.load_audio("transcribe.mp3")
    if params is not None and "language" in params:
        result = model.transcribe(audio, language=params['language'], verbose=False)
    else:
        result = model.transcribe(audio, verbose=False)
    filetype = result['language'] + ".txt"
    buffer = BytesIO(result["text"].encode())
    buffer.seek(0)
    return filetype, buffer
