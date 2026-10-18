`tone.mp3` and `silent.mp3` are original 0.6-second, 24 kHz mono test signals,
encoded locally with libmp3lame: a 440 Hz sine wave and zero PCM respectively.

`near-silent-provider.mp3` preserves the first failed public `light` adjective sample
from the OpenAI/Marin audition on 18 October 2026. It decodes successfully but has
almost no signal. The regression must reject it rather than redraw it to pass.
No competitor recordings or learner media are included.
