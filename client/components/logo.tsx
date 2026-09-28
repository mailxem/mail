
interface LogoProps {
  className?: string;
  onlyLogo?: boolean;
}

export function Logo({ className, onlyLogo }: LogoProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      xmlnsXlink="http://www.w3.org/1999/xlink"
      viewBox="0 0 50 50"
      width="100%"
      height="100%"
    >
      <g>
        <g>
          <line
            style={{
              fill: "none",
              stroke: "currentColor",
              strokeWidth: 2,
              strokeLinecap: "round",
              strokeMiterlimit: 10,
            }}
            x1="14.48"
            y1="37.74"
            x2="13.94"
            y2="43.01"
          />
          <path
            style={{
              fill: "none",
              stroke: "currentColor",
              strokeWidth: 2,
              strokeLinecap: "round",
              strokeMiterlimit: 10,
            }}
            d="M9,44.008c2-1,5-1,5-1h2"
          />
        </g>
      </g>
      <g>
        <g>
          <line
            style={{
              fill: "none",
              stroke: "currentColor",
              strokeWidth: 2,
              strokeLinecap: "round",
              strokeMiterlimit: 10,
            }}
            x1="21.57"
            y1="39.98"
            x2="20.94"
            y2="45.95"
          />
          <path
            style={{
              fill: "none",
              stroke: "currentColor",
              strokeWidth: 2,
              strokeLinecap: "round",
              strokeMiterlimit: 10,
            }}
            d="M16,47.008c2-1,5-1,5-1h2"
          />
        </g>
      </g>
      <path
        style={{
          fill: "none",
          stroke: "currentColor",
          strokeWidth: 2,
          strokeMiterlimit: 10,
        }}
        d="M26.156,14.406l3.258,7.804c-2.251,0.635-4.287,0.883-6.883-0.429L26.156,14.406z"
      />
      <path
        style={{
          fill: "none",
          stroke: "currentColor",
          strokeWidth: 2,
          strokeMiterlimit: 10,
        }}
        d="M34.25,7.417c4.25,0.083,6.167,4.833,4.333,8.375c-5.417-1.208-6.5-1.292-11.83-2.389C27.667,9.667,30.479,7.435,34.25,7.417z"
      />
      <line
        style={{
          fill: "none",
          stroke: "currentColor",
          strokeWidth: 2,
          strokeLinecap: "round",
          strokeMiterlimit: 10,
        }}
        x1="38.74"
        y1="23.721"
        x2="43.75"
        y2="28"
      />
      <line
        style={{
          fill: "none",
          stroke: "currentColor",
          strokeWidth: 2,
          strokeLinecap: "round",
          strokeMiterlimit: 10,
        }}
        x1="34.789"
        y1="26.818"
        x2="36.038"
        y2="29.557"
      />
      <line
        style={{
          fill: "none",
          stroke: "currentColor",
          strokeWidth: 2,
          strokeLinecap: "round",
          strokeMiterlimit: 10,
        }}
        x1="41.607"
        y1="19.733"
        x2="44.531"
        y2="20.449"
      />
      <line
        style={{
          fill: "none",
          stroke: "currentColor",
          strokeWidth: 2,
          strokeMiterlimit: 10,
        }}
        x1="25.083"
        y1="16.542"
        x2="21.167"
        y2="24.25"
      />
      <g>
        <path
          fill="currentColor"
          d="M26.962,41.256c-0.214,0.065-0.43,0.113-0.645,0.17l6.303,3.643c1.072,0.619,2.265,0.935,3.476,0.935c0.612,0,1.229-0.081,1.837-0.244c0.663-0.178,1.285-0.448,1.854-0.795c-0.814-0.027-1.609-0.144-2.357-0.383L26.962,41.256z"
        />
      </g>
      <path
        d="M36.211,33.749c-0.101,0.063-0.198,0.124-0.305,0.191c-1.716,1.076-2.961,1.852-3.364,2.04c-2.467,1.153-4.899,1.723-7.162,1.723c-3.161,0-5.991-1.111-8.122-3.297c-3.419-3.506-4.253-9.073-2.028-13.54c0.154-0.309,0.709-1.372,1.398-2.111c0.377-0.403,1.011-0.425,1.413-0.049c0.404,0.377,0.427,1.01,0.05,1.414c-0.53,0.568-1.017,1.531-1.07,1.639c-1.853,3.718-1.166,8.344,1.669,11.251c3.082,3.159,7.823,3.582,13.005,1.159c0.225-0.105,1.145-0.671,2.04-1.229l-2.102-0.687l-8.844-8.819L30.81,7.96L30.48,7.469c-1.688-2.521-4.277-4.041-7.506-4.395C22.736,3.052,21.791,3,21.5,3s-1.236,0.052-1.455,0.073C9.938,3.833,2.315,12.71,3.053,22.861c0.637,8.904,7.761,16.271,16.568,17.134c0.601,0.06,1.199,0.09,1.793,0.09c1.9,0,3.752-0.323,5.536-0.923l11.085,3.521c0.649,0.208,1.334,0.312,2.053,0.312c0.491,0,0.997-0.048,1.519-0.146c2.623-0.553,4.622-2.252,5.347-4.547l0.299-0.944L36.211,33.749z M22.5,12c0.828,0,1.5,0.672,1.5,1.5S23.328,15,22.5,15S21,14.328,21,13.5S21.672,12,22.5,12z"
        fill="currentColor"
      />
    </svg>
  );
}
