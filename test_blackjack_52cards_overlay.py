import subprocess
import json

# Standard 52-card deck for blackjack
suits = ['Hearts', 'Diamonds', 'Clubs', 'Spades']
ranks = ['2', '3', '4', '5', '6', '7', '8', '9', '10', 'J', 'Q', 'K', 'A']
cards = [f"{rank} of {suit}" for suit in suits for rank in ranks]

num_cards = 52

# Prepare the request for 52 random numbers
request = json.dumps({"nums": num_cards, "gamecode": "blackjack_test"})

cmd = [
    f"{subprocess.os.environ['USERPROFILE']}\\go\\bin\\grpcurl.exe",
    "-plaintext",
    "-import-path", "proto",
    "-proto", "proto/rng.proto",
    "-d", request,
    "localhost:6000",
    "sgc7pb.Rng/getRngs"
]

result = subprocess.run(cmd, capture_output=True, text=True)

try:
    response = json.loads(result.stdout)
    for card, rand_num in zip(cards, response["rngs"]):
        print(f"{card}: Random number assigned (overlay): {rand_num}")
except Exception:
    print("Raw output:", result.stdout)
    print("Error:", result.stderr)
