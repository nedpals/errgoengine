public class PrecisionLoss {
    public static void main(String[] args) {
        double largeNumber = 12345678901234567890.123456789;

        // Potential loss of precision: Found double, required float
        float smallNumber = largeNumber;
    }
}
