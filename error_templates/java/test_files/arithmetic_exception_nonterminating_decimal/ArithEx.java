import java.math.BigDecimal;

public class ArithEx {
    public static void main(String[] args) {
        BigDecimal a = new BigDecimal(5);
        BigDecimal b = new BigDecimal(3);
        BigDecimal result = a.divide(b);
        System.out.println(result);
    }
}
