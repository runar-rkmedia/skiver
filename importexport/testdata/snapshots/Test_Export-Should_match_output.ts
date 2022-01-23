/** Title: Project Foo Generated on: 2022-01-23 */

/**
 * British Fallbacks:
 *
 * - Eng British
 * - En British
 */
const British = {
  /** General Category */
  General: {
    /**
     * Welcoming the user to Foo
     *
     * # Variables (with example-values):
     *
     * ```yaml
     * userName: Rock
     * ```
     *
     * Value: Welcome, {{user}}
     */
    Welcome: 'General.Welcome',
    /** For use in forms */
    Forms: {
      /**
       * Buttons - You click them
       *
       * Surely you know what they are!
       */
      Buttons: {
        /**
         * The submit-button
         *
         * # Variables (with example-values):
         *
         * ```yaml
         * count: 42
         * ```
         *
         * Value: Go to checkout ({{count}})
         */
        GoToCheckout: 'General.Forms.Buttons.GoToCheckout',
      },
    },
  },
}

/**
 * Norwegian bokmål Fallbacks:
 *
 * - Nn-NO Norwegian Nynorsk
 * - No
 * - Dan
 * - Swe
 * - Eng British
 */
const NorwegianBokmål = {
  /** General Category */
  General: {
    /** For use in forms */
    Forms: {
      /**
       * Buttons - You click them
       *
       * Surely you know what they are!
       */
      Buttons: {
        /**
         * The submit-button
         *
         * # Variables (with example-values):
         *
         * ```yaml
         * count: 42
         * ```
         *
         * Value: Gå til utsjekk ({{count}})
         */
        GoToCheckout: 'General.Forms.Buttons.GoToCheckout',
      },
    },
  },
}
