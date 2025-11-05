# Piper TTS

A fast and local neural text-to-speech engine. This project uses the Piper C/C++ API.

Original project: [piper1-gpl](https://github.com/OHF-Voice/piper1-gpl)

## Install
```sh
git clone https://github.com/cemalgnlts/pipertts.git
./internal/libpiper/build.sh
export DYLD_LIBRARY_PATH="$PWD/internal/libpiper/install/lib:$DYLD_LIBRARY_PATH"
go run .
```

## Usage

### Download Models
Find and download the model for the language you will use here: [huggingface.co/rhasspy/piper-voices](https://huggingface.co/rhasspy/piper-voices/tree/main)

```sh
# download .onnx
curl -O https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/sam/medium/en_US-sam-medium.onnx

# download .onnx.json
curl -O https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/sam/medium/en_US-sam-medium.onnx.json

# move to models folder
mv en_US-sam-medium.onnx models
mv en_US-sam-medium.onnx.json models
```

### Run
```sh
go run . models/en_US-sam-medium.onnx 'hello world'
open output.wav
```

## License

[piper1-gpl](https://github.com/OHF-Voice/piper1-gpl/blob/main/COPYING)